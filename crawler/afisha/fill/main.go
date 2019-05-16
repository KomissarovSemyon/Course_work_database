package main

import (
	"database/sql"
	"flag"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

var (
	outDir string
)

const (
	repertoriesDir = "repertories"
	placesDir      = "places"
	scheduleDir    = "schedule"
)

var (
	cityLoader  *CityLoader
	tzLoader    *TZLoader
	placeLoader *PlaceLoader
	eventLoader *EventLoader
)

func main() {
	var connStr string

	flag.StringVar(&outDir, "out", "", "Crawl result output dir")
	flag.StringVar(&connStr, "conn", "", "Postgres connection specifier")
	doFillPlaces := flag.Bool("fill-places", false, "Fill places")
	doFillSessions := flag.String("fill-sessions", "", "Fill sessions for dates (comma separated)")

	flag.Parse()

	if outDir == "" {
		log.Fatal("out is required")
	}

	if connStr == "" {
		log.Fatal("conn is required")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to Open database", err)
	}

	cityLoader = NewCityLoader(db)
	tzLoader = NewTZLoader(db)
	placeLoader = NewPlaceLoader(db)
	eventLoader = NewEventLoader(db)

	if *doFillPlaces {
		err = fillPlaces(db)
		if err != nil {
			log.Fatal("fillPlaces failed:", err)
		}
	}

	if *doFillSessions != "" {
		dates := strings.Split(*doFillSessions, ",")
		for _, date := range dates {
			if err := fillSessions(db, date); err != nil {
				log.Fatalf("Failed to fill sessions for date: `%v` (%v)", date, err)
			}
		}
	}
}
