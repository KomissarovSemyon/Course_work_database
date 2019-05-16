package main

import (
	"database/sql"
)

type TZLoader struct {
	Loader
}

func NewTZLoader(db *sql.DB) *TZLoader {
	return &TZLoader{
		Loader: *NewLoader(db, LoaderConfig{
			Table:     "timezones",
			FieldName: "name",
			FieldID:   "timezone_id",
		}),
	}
}
