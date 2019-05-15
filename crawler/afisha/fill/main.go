package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"path"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stek29/kr/crawler/afisha"
	"github.com/stek29/kr/crawler/afisha/util"
)

var (
	outDir string
)

const (
	// repertoriesDir = "repertories"
	placesDir = "places"
	// scheduleDir    = "schedule"
)

// XXX: duplicated in crawl
func loadPlaces() ([]afisha.Place, error) {
	placeFiles, err := filepath.Glob(path.Join(outDir, placesDir, "*"))
	if err != nil {
		return nil, errors.Wrap(err, "Place files Glob failed: ")
	}
	log.Printf("Loading places from %v files", len(placeFiles))

	var places []afisha.Place

	for i, fn := range placeFiles {
		log.Printf("INFO: Loading place file #%d (%v)", i, fn)
		var chunk []afisha.Place
		if err := util.UnmarshalFromFile(fn, &chunk); err != nil {
			log.Printf("Failed to load file, skipping: %v", err)
			continue
		}

		places = append(places, chunk...)
	}

	return places, nil
}

var (
	cityLoader  *CityLoader
	tzLoader    *TZLoader
	placeLoader *PlaceLoader
)

func fillPlaces(db *sql.DB) error {
	places, err := loadPlaces()
	if err != nil {
		return err
	}

	citymap := map[string]afisha.City{}
	tzset := map[string]struct{}{}
	for _, pl := range places {
		if _, ok := citymap[pl.City.Name]; !ok {
			citymap[pl.City.Name] = pl.City
			tzset[pl.City.TimeZone] = struct{}{}
		}
	}

	tzs := make([]string, len(tzset))
	i := 0
	for tz := range tzset {
		tzs[i] = tz
		i++
	}

	tzmap, err := tzLoader.TimeZoneIDs(tzs)
	if err != nil {
		return err
	}

	if len(tzmap) < len(tzs) {
		log.Printf("Expected to get %d timezones, but got %d", len(tzs), len(tzmap))
		for _, tz := range tzs {
			if _, ok := tzmap[tz]; !ok {
				log.Printf("TZ missing: %s", tz)
			}
		}
		return errors.Errorf("Cant find some timezones")
	}

	cities := make([]CityData, len(citymap))
	i = 0
	for _, city := range citymap {
		var ok bool

		cities[i].CountryCode = [...]byte{'R', 'U'}
		cities[i].Name = city.Name
		cities[i].YaName = city.ID
		cities[i].TimeZoneID, ok = tzmap[city.TimeZone]
		if !ok {
			panic(errors.Errorf("Unexpected cache miss for tz: %v", city.TimeZone))
		}
		i++
	}

	cityIDmap, err := cityLoader.CityIDsCreating(cities)
	if err != nil {
		return err
	}

	var placeDatas []PlaceData
	maxl := 0
	maxt := ""
	for _, pl := range places {
		cid, ok := cityIDmap[pl.City.ID]
		if !ok {
			panic(errors.Errorf("Unexpected cache miss for city: %v", pl.City))
		}

		placeDatas = append(placeDatas, PlaceData{
			CityID:  cid,
			Name:    pl.Title,
			Address: pl.Address,
			Long:    pl.Coordinates.Longitude,
			Lat:     pl.Coordinates.Latitude,
			YaID:    pl.ID,
		})

		if len(pl.Address) > maxl {
			maxl = len(pl.Address)
			maxt = pl.Address
		}
	}
	fmt.Printf("longest addr: %d (%s)", maxl, maxt)

	const chunkSize = 100

	for i := 0; i < len(placeDatas); i += chunkSize {
		end := i + chunkSize

		if end > len(placeDatas) {
			end = len(placeDatas)
		}

		chunk := placeDatas[i:end]
		// log.Printf("chunk: %v", chunk)

		_, err := placeLoader.PlaceIDsCreating(chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var connStr string

	flag.StringVar(&outDir, "out", "", "Crawl result output dir")
	flag.StringVar(&connStr, "conn", "", "Postgres connection specifier")

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

	err = fillPlaces(db)
	if err != nil {
		panic(err)
	}

	// loader := cityLoader
	// fmt.Println("City ID by Name `moscow`:")
	// fmt.Println(loader.CityID("moscow"))
	// fmt.Println("City ID by Name `moscow`:")
	// fmt.Println(loader.CityID("moscow"))

	// fmt.Println("City IDs by Names: `moscow`:")
	// fmt.Println(loader.CityIDs([]string{"moscow"}))

	// fmt.Println("Cities (ensured): moscow, abakan:")
	// fmt.Println(loader.CityIDsCreating([]CityData{
	// 	CityData{
	// 		[...]byte{'R', 'U'},
	// 		"Москва",
	// 		"moscow",
	// 		1,
	// 	},
	// 	CityData{
	// 		[...]byte{'R', 'U'},
	// 		"Абакан",
	// 		"abakan",
	// 		1,
	// 	},
	// }))
}
