package main

import (
	"database/sql"
)

type CityLoader struct {
	Loader
}

func NewCityLoader(db *sql.DB) *CityLoader {
	return &CityLoader{
		Loader: *NewLoader(db, LoaderConfig{
			Table:     "cities",
			FieldName: "ya_name",
			FieldID:   "city_id",
		}),
	}
}

type CityDataItem struct {
	CountryCode [2]byte
	Name        string
	YaName      string
	TimeZoneID  int
}

type CityData []CityDataItem

func (CityData) Fields() []string {
	return []string{"country_code", "name", "ya_name", "timezone_id"}
}

func (CityData) InsertFormat() (string, int) {
	return "$%d, $%d, $%d, $%d", 4
}

func (d CityData) Names() []string {
	res := make([]string, len(d))
	for i := range d {
		res[i] = d[i].YaName
	}
	return res
}

func (d CityData) Values(filter map[string]struct{}) []interface{} {
	var res []interface{}

	for _, item := range d {
		if _, ok := filter[item.YaName]; !ok {
			continue
		}

		var idata [4]interface{}

		idata[0] = string(item.CountryCode[:])
		idata[1] = item.Name
		idata[2] = item.YaName
		idata[3] = item.TimeZoneID

		res = append(res, idata[:]...)
	}

	return res
}
