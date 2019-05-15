package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type PlaceLoader struct {
	cache map[string]int
	mu    sync.RWMutex
	db    *sql.DB
}

func NewPlaceLoader(db *sql.DB) *PlaceLoader {
	return &PlaceLoader{
		cache: make(map[string]int),
		db:    db,
	}
}

func (l *PlaceLoader) PlaceID(yaID string) (int, error) {
	l.mu.RLock()
	if id, ok := l.cache[yaID]; ok {
		l.mu.RUnlock()
		return id, nil
	}
	l.mu.RUnlock()

	row := l.db.QueryRow("SELECT cinema_id FROM cinemas WHERE ya_id = $1", yaID)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	l.mu.Lock()
	l.cache[yaID] = id
	l.mu.Unlock()

	return id, nil
}

func (l *PlaceLoader) PlaceIDs(yaIDs []string) (map[string]int, error) {
	results := map[string]int{}

	// Only unique names
	uids := make(map[string]struct{})
	for _, yaID := range yaIDs {
		uids[yaID] = struct{}{}
	}

	var missIDs []string
	l.mu.RLock()
	for yaID := range uids {
		if id, ok := l.cache[yaID]; ok {
			results[yaID] = id
		} else {
			missIDs = append(missIDs, yaID)
		}
	}
	l.mu.RUnlock()

	if len(missIDs) == 0 {
		return results, nil
	}

	parts := make([]string, len(missIDs))
	inames := make([]interface{}, len(missIDs))

	for i := 0; i != len(missIDs); i++ {
		parts[i] = "$" + strconv.Itoa(i+1)
		inames[i] = missIDs[i]
	}

	statement := "SELECT cinema_id, ya_id FROM cinemas WHERE ya_id IN (" + strings.Join(parts, ",") + ")"

	rows, err := l.db.Query(statement, inames...)

	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for rows.Next() {
		var id int
		var yaID string

		if err := rows.Scan(&id, &yaID); err != nil {
			return nil, err
		}

		results[yaID] = id
		l.cache[yaID] = id
	}

	return results, nil
}

type PlaceData struct {
	CityID  int
	Name    string
	Address string
	Lat     float32
	Long    float32
	YaID    string
}

// XXX: CopyIn?
func (l *PlaceLoader) InsertPlaces(places []PlaceData) (map[string]int, error) {
	if len(places) == 0 {
		return map[string]int{}, nil
	}

	parts := make([]string, len(places))
	idata := make([]interface{}, len(places)*6)

	for i, place := range places {
		ii := i * 6
		parts[i] = fmt.Sprintf("( $%d, $%d, point($%d, $%d), $%d, $%d)", ii+1, ii+2, ii+3, ii+4, ii+5, ii+6)

		idata[ii+0] = place.Name
		idata[ii+1] = place.Address
		idata[ii+2] = place.Lat
		idata[ii+3] = place.Long
		idata[ii+4] = place.CityID
		idata[ii+5] = place.YaID
	}

	statement := "INSERT INTO cinemas (name, address, loc, city_id, ya_id) VALUES " +
		strings.Join(parts, ",") +
		" RETURNING cinema_id, ya_id"

	rows, err := l.db.Query(statement, idata...)

	if err != nil {
		return nil, err
	}

	results := map[string]int{}
	l.mu.Lock()
	defer l.mu.Unlock()

	for rows.Next() {
		var id int
		var yaID string

		if err := rows.Scan(&id, &yaID); err != nil {
			return nil, err
		}

		results[yaID] = id
		l.cache[yaID] = id
	}

	return results, nil
}

func (l *PlaceLoader) PlaceIDsCreating(places []PlaceData) (map[string]int, error) {
	yaIDs := make([]string, len(places))
	for i, place := range places {
		yaIDs[i] = place.YaID
	}

	result, err := l.PlaceIDs(yaIDs)
	if err != nil {
		return nil, err
	}

	createThese := map[string]*PlaceData{}
	for i := 0; i != len(places); i++ {
		pl := &places[i]
		if _, ok := result[pl.YaID]; !ok {
			createThese[pl.YaID] = pl
		}
	}

	if len(createThese) == 0 {
		return result, nil
	}

	insertPlaces := make([]PlaceData, len(createThese))
	i := 0
	for _, pl := range createThese {
		insertPlaces[i] = *pl
		i++
	}

	inserted, err := l.InsertPlaces(insertPlaces)
	if err != nil {
		return nil, err
	}

	for name, id := range inserted {
		result[name] = id
	}

	return result, nil
}
