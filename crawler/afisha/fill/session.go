package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stek29/kr/crawler/afisha"
	"github.com/stek29/kr/crawler/afisha/util"
)

const (
	SessionType3D   = 1
	SessionTypeIMAX = 2
)

type Session struct {
	Hall     string
	CinemaID int
	CityID   int
	EventID  string
	MovieID  int
	Type     int
	YaID     string
	Date     time.Time
	PriceMin int
	PriceMax int
}

func (s *Session) UniqueKey() string {
	if s.YaID != "" {
		return s.YaID
	}

	return fmt.Sprintf("%v;%d;%s;%v", s.Hall, s.CinemaID, s.EventID, s.Date)
}

func InsertSessions(db *sql.DB, sessions []Session) error {
	txn, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn(
		"sessions",
		"hall_name",
		"cinema_id",
		"city_id",
		"movie_id",
		"type",
		"ya_id",
		"date",
		"price_min",
		"price_max",
	))

	if err != nil {
		return err
	}

	for _, sess := range sessions {
		var hall, yaID interface{}

		if sess.Hall != "" {
			hall = sess.Hall
		}
		if sess.YaID != "" {
			yaID = sess.YaID
		}

		_, err = stmt.Exec(
			hall,
			sess.CinemaID,
			sess.CityID,
			sess.MovieID,
			sess.Type,
			yaID,
			sess.Date,
			sess.PriceMin/100,
			sess.PriceMax/100,
		)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	err = txn.Commit()
	if err != nil {
		return err
	}

	return nil
}

func loadSessions(date afisha.Date) ([]Session, error) {
	sessionFiles, err := filepath.Glob(path.Join(outDir, scheduleDir, date.String(), "*", "*.json"))
	if err != nil {
		return nil, errors.Wrap(err, "Session files Glob failed: ")
	}
	log.Printf("Loading sessions from %d files", len(sessionFiles))

	parseSessionFilename := func(fn string) (city, placeID string) {
		pathItems := strings.Split(fn, string(os.PathSeparator))

		placeID = pathItems[len(pathItems)-1]
		placeID = strings.Replace(placeID, ".json", "", 1)

		city = pathItems[len(pathItems)-2]
		return
	}

	var cities []string
	var placeIDs []string

	for _, fn := range sessionFiles {
		city, placeID := parseSessionFilename(fn)
		cities = append(cities, city)
		placeIDs = append(placeIDs, placeID)
	}

	placeMap, err := placeLoader.GetIDs(placeIDs)
	if err != nil {
		return nil, err
	}
	log.Printf("%d places loaded", len(placeMap))

	cityMap, err := cityLoader.GetIDs(cities)
	if err != nil {
		return nil, err
	}
	log.Printf("%d cities loaded", len(cityMap))

	var sessions []Session

	yaEvents := map[string]yaEventInfo{}

	antiDupe := map[string]struct{}{}

	for _, fn := range sessionFiles {
		var items []afisha.ScheduleItem

		if err := util.UnmarshalFromFile(fn, &items); err != nil {
			log.Printf("Failed to load file %v, skipping: %v", fn, err)
			continue
		}
		log.Printf("%d items for %s", len(items), fn)

		city, placeYaID := parseSessionFilename(fn)
		placeID, ok := placeMap[placeYaID]
		if !ok {
			log.Fatalf("Unexpected placeMap cache miss: %s", placeYaID)
		}
		cityID, ok := cityMap[city]
		if !ok {
			log.Fatalf("Unexpected cityMap cache miss: %s", city)
		}

		for _, item := range items {
			for _, sched := range item.Schedule {
				eventID := item.Event.ID

				kpID := kpIDFromURL(item.Event.Kinopoisk.URL)
				if info, ok := yaEvents[eventID]; !ok || (info.kpID == 0 && kpID != 0) {
					yaEvents[eventID] = yaEventInfo{
						kpID: kpID,
						url:  item.Event.URL,
					}
				}

				fmt := strings.ToLower(string(sched.Format))

				sessType := 0
				if strings.Contains(fmt, "3d") {
					sessType |= SessionType3D
				}
				if strings.Contains(fmt, "imax") {
					sessType |= SessionTypeIMAX
				}

				for _, sess := range sched.Sessions {
					var ticketID string
					if tid := sess.Ticket.ID; tid != "" {
						ticketIDBytes, err := base64.StdEncoding.DecodeString(tid)
						if err != nil {
							log.Printf("Failed to decode ticketID (%s) for event=%s city=%s place=%s, skipping: %v", tid, item.Event.ID, city, placeYaID, err)
							continue
						}
						ticketID = string(ticketIDBytes)
					}

					dateTime, err := time.Parse(afisha.DateTimeLayout, sess.Datetime)
					if err != nil {
						log.Printf("Failed to parse date (%s) for event=%s city=%s place=%s, skipping: %v", sess.Datetime, item.Event.ID, city, placeYaID, err)
						continue
					}

					session := Session{
						Hall:     sess.HallName,
						CinemaID: placeID,
						CityID:   cityID,
						EventID:  eventID,
						Type:     sessType,
						YaID:     ticketID,
						Date:     dateTime,
						PriceMin: sess.Ticket.Price.Min,
						PriceMax: sess.Ticket.Price.Max,
					}

					antiDupeKey := session.UniqueKey()
					if _, ok := antiDupe[antiDupeKey]; ok {
						log.Printf("Duplicate session detected, skipping: %v", antiDupeKey)
					} else {
						antiDupe[antiDupeKey] = struct{}{}
						sessions = append(sessions, session)
					}
				}
			}
		}
	}

	eventMap, err := loadEvents(yaEvents)
	if err != nil {
		log.Printf("Failed to load events!")
		return nil, err
	}

	for i := range sessions {
		var ok bool
		sessions[i].MovieID, ok = eventMap[sessions[i].EventID]
		if !ok {
			log.Fatalf("Unexpected eventMap cache miss (%s)", sessions[i].EventID)
		}
	}

	return sessions, nil
}

func fillSessions(db *sql.DB, dateStr string) error {
	date, err := afisha.ParseDate(dateStr)
	if err != nil {
		return err
	}
	sessions, err := loadSessions(date)
	if err != nil {
		return err
	}

	log.Printf("Loaded %d sessions for date %s", len(sessions), date)

	const chunkSize = 1000
	log.Printf("Saving %d sessions in chunks of %d", len(sessions), chunkSize)
	for i := 0; i < len(sessions); i += chunkSize {
		end := i + chunkSize

		if end > len(sessions) {
			end = len(sessions)
		}

		chunk := sessions[i:end]

		err := InsertSessions(db, chunk)
		if err != nil {
			return err
		}
	}
	return nil
}
