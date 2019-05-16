package main

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/stek29/kr/crawler/afisha"
	"github.com/stek29/kr/crawler/afisha/util"
)

type EventLoader struct {
	Loader
}

func NewEventLoader(db *sql.DB) *EventLoader {
	return &EventLoader{
		Loader: *NewLoader(db, LoaderConfig{
			Table:     "movies",
			FieldName: "ya_event_id",
			FieldID:   "movie_id",
		}),
	}
}

type EventDataItem struct {
	EventID   string
	KpID      int
	TitleRU   string
	TitleEN   string
	Year      int
	Duration  int
	ReleaseRU afisha.Date
	// KpRating       int
	AgeRestriction int
	CountryCode    string
}

type EventData []EventDataItem

func (EventData) Fields() []string {
	return []string{"ya_event_id", "kp_id", "title_ru", "title_en", "year", "duration", "release_ru", "age_restriction", "country_code"}
}

func (EventData) InsertFormat() (string, int) {
	return "$%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d", 9
}

func (d EventData) Names() []string {
	res := make([]string, len(d))
	for i := range d {
		res[i] = d[i].EventID
	}
	return res
}

func (d EventData) Values(filter map[string]struct{}) []interface{} {
	var res []interface{}

	for _, item := range d {
		if _, ok := filter[item.EventID]; !ok {
			continue
		}

		var idata [9]interface{}

		idata[0] = item.EventID
		if item.KpID != 0 {
			idata[1] = item.KpID
		}
		if item.TitleRU != "" {
			idata[2] = item.TitleRU
		}
		if item.TitleEN != "" {
			idata[3] = item.TitleEN
		}
		if item.Year != 0 {
			idata[4] = item.Year
		}
		if item.Duration != 0 {
			idata[5] = item.Duration
		}
		if !item.ReleaseRU.IsZero() {
			idata[6] = item.ReleaseRU.String()
		}
		if item.AgeRestriction != 0 {
			idata[7] = item.AgeRestriction
		}
		if item.CountryCode != "" {
			idata[8] = item.CountryCode
		}

		res = append(res, idata[:]...)
	}

	return res
}

var ruMonths = []struct {
	match string
	val   time.Month
}{
	{"январ", time.January},
	{"феврал", time.February},
	{"март", time.March},
	{"апрел", time.April},
	{"июн", time.June},
	{"июл", time.July},
	{"август", time.August},
	{"сентябр", time.September},
	{"октябр", time.October},
	{"ноябр", time.November},
	{"декабр", time.December},
	{"ма", time.May},
}

func parseRUDate(value string) (afisha.Date, error) {
	dateParts := strings.SplitN(value, " ", 3)
	if len(dateParts) != 3 {
		return afisha.Date{}, errors.New("Cant split in 3 parts")
	}

	day, err := strconv.Atoi(dateParts[0])
	if err != nil {
		return afisha.Date{}, errors.Wrapf(err, "Failed to parse day %s", dateParts[0])
	}

	year, err := strconv.Atoi(dateParts[2])
	if err != nil {
		return afisha.Date{}, errors.Wrapf(err, "Failed to parse year %s", dateParts[2])
	}

	var month time.Month
	monthStr := dateParts[1]
	matched := false

	for _, def := range ruMonths {
		if strings.Contains(monthStr, def.match) {
			month = def.val
			matched = true
		}
	}

	if !matched {
		return afisha.Date{}, errors.Errorf("Failed to parse month %s", monthStr)
	}

	return afisha.MakeDate(year, month, day), nil
}

var durationRegexp = regexp.MustCompile(`(\d+) мин.`)

var kpRegexp = regexp.MustCompile(`kinopoisk.ru/film/(\d+)`)

func fetchAfisha(evID, afishaURL string) (*EventDataItem, error) {
	req, err := http.NewRequest("GET", "https://afisha.yandex.ru/"+afishaURL, nil)
	if err != nil {
		return nil, err
	}
	// fuck captcha
	req.Header.Set("Cookie", "bltsr=1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if strings.Contains(string(data), "captcha") {
		return nil, errors.New("CAPTCHA")
	}

	var item EventDataItem

	item.EventID = evID

	{
		match := kpRegexp.FindSubmatch(data)
		if len(match) == 2 {
			kpID, err := strconv.Atoi(string(match[1]))
			if err == nil {
				item.KpID = kpID
			}
		}
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	getText := func(sel *goquery.Selection, selector string) string {
		return strings.TrimSpace(sel.Find(selector).Text())
	}

	item.AgeRestriction, _ = strconv.Atoi(strings.Trim(getText(doc.Selection, `[class=event-heading__content-rating]`), "+"))
	item.TitleRU = getText(doc.Selection, `[class="event-heading__title"]`)

	doc.Find(`[class="event-attributes__row"]`).Each(func(_ int, sel *goquery.Selection) {
		key := getText(sel, `[class="event-attributes__category"]`)
		value := getText(sel, `[class="event-attributes__category-value"]`)

		switch key {
		case "Оригинальное название":
			item.TitleEN = value
		case "Год производства":
			item.Year, _ = strconv.Atoi(value)
		case "Время":
			match := durationRegexp.FindSubmatch([]byte(value))
			if len(match) != 2 {
				log.Printf("Failed to parse (%v: %v): durationRegexp didn't match", key, value)
				break
			}

			item.Duration, _ = strconv.Atoi(string(match[1]))
		case "Страна":
			// item.CountryCode
		case "Премьера":
			item.ReleaseRU, err = parseRUDate(value)
			if err != nil {
				log.Printf("Failed to parse (%v: %v): %v", key, value, err)
			}

		case "Режиссёр", "Композитор", "В ролях":

		default:
			log.Printf("Unknown category: %v", key)
		}
	})

	return &item, nil
}

type yaEventInfo struct {
	kpID int
	url  string
}

// loadRepertoryKpIDs loads eventID=>kpID from repertories
func loadRepertoryKpIDs() (map[string]int, error) {
	repertoryFiles, err := filepath.Glob(path.Join(outDir, repertoriesDir, "*.json"))
	if err != nil {
		return nil, errors.Wrap(err, "Repertory files Glob failed: ")
	}
	log.Printf("Loading repertories from %d files", len(repertoryFiles))

	result := map[string]int{}

	for _, fn := range repertoryFiles {
		var items []afisha.RepertoryItem
		err := util.UnmarshalFromFile(fn, &items)
		if err != nil {
			log.Printf("Failed to load file %v, skipping: %v", fn, err)
			continue
		}

		for _, item := range items {
			evID := item.Event.ID
			kpID := kpIDFromURL(item.Event.Kinopoisk.URL)

			if kpID == 0 {
				continue
			}

			if _, ok := result[evID]; ok {
				continue
			}

			result[evID] = kpID
		}
	}

	return result, nil
}

func loadEvents(events map[string]yaEventInfo) (map[string]int, error) {
	var err error
	log.Printf("Found %d events", len(events))

	names := make([]string, len(events))
	i := 0
	for evID := range events {
		names[i] = evID
		i++
	}

	result, err := eventLoader.GetIDs(names)
	if err != nil {
		return nil, err
	}

	kpIDMap, err := loadRepertoryKpIDs()
	if err != nil {
		return nil, err
	}

	i = 0
	for evID, info := range events {
		if info.kpID != 0 {
			continue
		}

		if kpID, ok := kpIDMap[evID]; ok {
			info.kpID = kpID
			events[evID] = info
		} else {
			i++
		}
	}

	var createEvents EventData

	for evID, info := range events {
		if _, ok := result[evID]; ok {
			continue
		}

		item := &EventDataItem{
			EventID: evID,
			KpID:    info.kpID,
		}

		if info.kpID == 0 {
			log.Printf("kpID not found for event (%s) and isn't in repertory, fetching by url (%s)", evID, info.url)
			item, err = fetchAfisha(evID, info.url)
			if err != nil {
				log.Printf("Failed to fetch kp id (skipping event %s!): %v", evID, err)
				continue
			}

			// Save it in case we ever need it
			info.kpID = item.KpID
			events[evID] = info
		}

		createEvents = append(createEvents, *item)
	}

	if len(createEvents) == 0 {
		return result, err
	}

	created, err := eventLoader.GetIDsCreating(createEvents)
	if err != nil {
		return nil, err
	}

	for name, id := range created {
		result[name] = id
	}

	return result, err
}

func kpIDFromURL(url string) int {
	if url == "" {
		return 0
	}

	kpData := strings.Split(url, "/")
	kpID, err := strconv.Atoi(kpData[len(kpData)-1])
	if err != nil {
		panic(err)
	}
	return kpID
}
