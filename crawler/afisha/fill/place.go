package main

import (
	"database/sql"
	"log"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stek29/kr/crawler/afisha"
	"github.com/stek29/kr/crawler/afisha/util"
)

type PlaceLoader struct {
	Loader
}

func NewPlaceLoader(db *sql.DB) *PlaceLoader {
	return &PlaceLoader{
		Loader: *NewLoader(db, LoaderConfig{
			Table:     "cinemas",
			FieldName: "ya_id",
			FieldID:   "cinema_id",
		}),
	}
}

type PlaceDataItem struct {
	CityID  int
	Name    string
	Address string
	Lat     float32
	Long    float32
	YaID    string
}

type PlaceData []PlaceDataItem

func (PlaceData) Fields() []string {
	return []string{"name", "address", "loc", "city_id", "ya_id"}
}

func (PlaceData) InsertFormat() (string, int) {
	return "$%d, $%d, point($%d, $%d), $%d, $%d", 6
}

func (d PlaceData) Names() []string {
	res := make([]string, len(d))
	for i := range d {
		res[i] = d[i].YaID
	}
	return res
}

func (d PlaceData) Values(filter map[string]struct{}) []interface{} {
	var res []interface{}

	for _, item := range d {
		if _, ok := filter[item.YaID]; !ok {
			continue
		}

		var idata [6]interface{}

		idata[0] = item.Name
		idata[1] = item.Address
		idata[2] = item.Lat
		idata[3] = item.Long
		idata[4] = item.CityID
		idata[5] = item.YaID

		res = append(res, idata[:]...)
	}

	return res
}

// XXX: duplicated in crawl
func loadPlaces() ([]afisha.Place, error) {
	placeFiles, err := filepath.Glob(path.Join(outDir, placesDir, "*"))
	if err != nil {
		return nil, errors.Wrap(err, "Place files Glob failed: ")
	}
	log.Printf("Loading places from %v files", len(placeFiles))

	var places []afisha.Place

	for _, fn := range placeFiles {
		var chunk []afisha.Place
		if err := util.UnmarshalFromFile(fn, &chunk); err != nil {
			log.Printf("Failed to load file %v, skipping: %v", fn, err)
			continue
		}

		places = append(places, chunk...)
	}

	return places, nil
}

func fillPlaces(db *sql.DB) error {
	places, err := loadPlaces()
	if err != nil {
		return err
	}
	log.Printf("Loaded %d places", len(places))

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

	tzmap, err := tzLoader.GetIDs(tzs)
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

	log.Printf("Loaded %d timezones", len(tzmap))

	cities := make(CityData, len(citymap))
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

	log.Printf("Saving %d cities", len(cities))
	cityIDmap, err := cityLoader.GetIDsCreating(cities)
	if err != nil {
		return err
	}

	var placeDatas PlaceData

	for _, pl := range places {
		cid, ok := cityIDmap[pl.City.ID]
		if !ok {
			panic(errors.Errorf("Unexpected cache miss for city: %v", pl.City))
		}

		placeDatas = append(placeDatas, PlaceDataItem{
			CityID:  cid,
			Name:    pl.Title,
			Address: pl.Address,
			Long:    pl.Coordinates.Longitude,
			Lat:     pl.Coordinates.Latitude,
			YaID:    pl.ID,
		})
	}

	const chunkSize = 100
	log.Printf("Saving %d places in chunks of %d", len(placeDatas), chunkSize)
	for i := 0; i < len(placeDatas); i += chunkSize {
		end := i + chunkSize

		if end > len(placeDatas) {
			end = len(placeDatas)
		}

		chunk := placeDatas[i:end]

		_, err := placeLoader.GetIDsCreating(chunk)
		if err != nil {
			return err
		}
	}

	return nil
}
