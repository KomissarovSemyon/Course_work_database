package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/stek29/kr/crawler/afisha"
	"github.com/stek29/kr/crawler/afisha/util"
)

const (
	repertoriesDir = "repertories"
	placesDir      = "places"
	scheduleDir    = "schedule"
)

var (
	cities []string
	outDir string
)

func crawlCityRepertories() error {
	log.Println("Crawling repertories by Cities")

	repertoriesPath := path.Join(outDir, repertoriesDir)
	if err := os.MkdirAll(repertoriesPath, 0755); err != nil {
		return errors.Wrap(err, "Failed to prepare repertories dir")
	}

	for _, city := range cities {
		params := afisha.RepertoryParams{
			Limit:  20,
			Offset: 0,
			City:   city,
		}

		allEvents, err := afisha.GetRepetoryFull(http.DefaultClient, &params)
		if err != nil {
			log.Printf("WARN: Failed to get repertory for %v city, skipping: %v", city, err)
			continue
		}

		fn := path.Join(repertoriesPath, city+".json")
		err = util.MarshalIntoFile(fn, allEvents.Data)
		if err != nil {
			log.Printf("WARN: failed to save %v city repertory, skipping: %v", city, err)
			continue
		}
	}

	return nil
}

func crawlPlaces() error {
	log.Println("Crawling repertories by Cities")

	placesPath := path.Join(outDir, placesDir)
	if err := os.MkdirAll(placesPath, 0755); err != nil {
		return errors.Wrap(err, "Failed to prepare places dir")
	}

	for _, city := range cities {
		params := afisha.PlacesParams{
			Limit:  20,
			Offset: 0,
			City:   city,
		}

		allPlaces, err := afisha.GetPlacesFull(http.DefaultClient, &params)
		if err != nil {
			log.Printf("WARN: Failed to get places for %v city, skipping: %v", city, err)
			continue
		}

		fn := path.Join(placesPath, city+".json")
		err = util.MarshalIntoFile(fn, allPlaces.Items)
		if err != nil {
			log.Printf("WARN: failed to save %v city places, skipping: %v", city, err)
			continue
		}
	}

	return nil
}

type placeInfo struct {
	placeID string
	title   string
	city    string
}

func loadPlaces() ([]placeInfo, error) {
	placeFiles, err := filepath.Glob(path.Join(outDir, placesDir, "*"))
	if err != nil {
		return nil, errors.Wrap(err, "Place files Glob failed: ")
	}
	log.Printf("Loading places from %v files", len(placeFiles))

	var places []placeInfo

	for i, fn := range placeFiles {
		log.Printf("INFO: Loading place file #%d (%v)", i, fn)
		var chunk []afisha.Place
		if err := util.UnmarshalFromFile(fn, &chunk); err != nil {
			log.Printf("Failed to load file, skipping: %v", err)
			continue
		}

		for _, pl := range chunk {
			places = append(places, placeInfo{
				placeID: pl.ID,
				title:   pl.Title,
				city:    pl.City.ID,
			})
		}
	}

	return places, nil
}

func crawlPlaceSchedules(dateStr string) error {
	date, err := afisha.ParseDate(dateStr)
	if err != nil {
		return errors.Wrapf(err, "Invalid date for do-place-schedules: `%v`", dateStr)
	}

	places, err := loadPlaces()
	if err != nil {
		return errors.Wrap(err, "Failed to load places")
	}

	log.Printf("Loaded %v places", len(places))

	schedulesPath := path.Join(outDir, scheduleDir, date.String())
	if err := os.MkdirAll(schedulesPath, 0755); err != nil {
		return errors.Wrap(err, "Failed to prepare schedules dir")
	}

	for i, pl := range places {
		log.Printf("INFO: Processing place %d/%d (%s - %s from city %s)", i+1, len(places), pl.placeID, pl.title, pl.city)

		outPath := path.Join(schedulesPath, pl.city)
		if err := os.MkdirAll(outPath, 0755); err != nil {
			log.Printf("WARN: Failed to prepare schedules dir for places of city %v, skipping place %v: %v", pl.city, pl.placeID, err)
			continue
		}

		params := afisha.ScheduleCinemaParams{
			PlaceID: pl.placeID,
			City:    pl.city,
			Date:    date,
			Limit:   20,
		}

		schd, err := afisha.GetScheduleCinemaFull(http.DefaultClient, &params)
		if err != nil {
			log.Printf("WARN: Failed to load schedules for place %v (city=%v), skipping: %v", pl.placeID, pl.city, err)
			continue
		}

		util.MarshalIntoFile(path.Join(outPath, pl.placeID+".json"), schd.Items)
		if err != nil {
			log.Printf("WARN: Failed to save schedules for place %v (city=%v), skipping: %v", pl.placeID, pl.city, err)
			continue
		}
	}

	return nil
}

func main() {
	cityListFile := flag.String("city-list", "", "City list in JSON")

	doCityRepertories := flag.Bool("do-city-repertories", false, "Crawl repertories by city")
	doPlaces := flag.Bool("do-places", false, "Crawl places by city")

	doPlaceSchedules := flag.String("do-place-schedules", "", "Crawl schedule for all places for specific date")

	flag.StringVar(&outDir, "out", "", "Crawl result output dir")
	flag.Parse()

	if outDir == "" {
		log.Fatal("-out is required")
	}

	if *cityListFile != "" {
		log.Println("Loading City List")
		if err := util.UnmarshalFromFile(*cityListFile, &cities); err != nil {
			log.Fatalf("Failed to load city list from %v: %v", *cityListFile, err)
		}
		log.Printf("Loaded %v cities", len(cities))
	} else if *doCityRepertories || *doPlaces {
		log.Fatal("City list is required for do-city-repertories/do-places")
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal("Failed to prepare output dir", err)
	}

	log.Printf("Prepared output dir")

	if *doCityRepertories {
		if err := crawlCityRepertories(); err != nil {
			log.Fatalf("Failed to crawl city repertories: %v", err)
		}
	}

	if *doPlaces {
		if err := crawlPlaces(); err != nil {
			log.Fatalf("Failed to crawl city places: %v", err)
		}
	}

	if plsDate := *doPlaceSchedules; plsDate != "" {
		dates := strings.Split(plsDate, ",")
		for _, date := range dates {
			if err := crawlPlaceSchedules(date); err != nil {
				log.Fatalf("Failed to crawl date: `%v` (%v)", date, err)
			}
		}
	}
}
