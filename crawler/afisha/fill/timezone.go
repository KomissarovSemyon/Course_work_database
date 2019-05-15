package main

import (
	"database/sql"
	"strconv"
	"strings"
	"sync"
)

type TZLoader struct {
	cache map[string]int
	mu    sync.RWMutex
	db    *sql.DB
}

func NewTZLoader(db *sql.DB) *TZLoader {
	return &TZLoader{
		cache: make(map[string]int),
		db:    db,
	}
}

var ErrNotFound = sql.ErrNoRows

func (l *TZLoader) TimeZoneID(name string) (int, error) {
	l.mu.RLock()
	if id, ok := l.cache[name]; ok {
		l.mu.RUnlock()
		return id, nil
	}
	l.mu.RUnlock()
	row := l.db.QueryRow("SELECT timezone_id FROM timezones WHERE name = $1", name)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	l.mu.Lock()
	l.cache[name] = id
	l.mu.Unlock()

	return id, nil
}

func (l *TZLoader) TimeZoneIDs(names []string) (map[string]int, error) {
	results := map[string]int{}

	var missNames []string

	// Only unique names
	unames := make(map[string]struct{})
	for _, name := range names {
		unames[name] = struct{}{}
	}

	l.mu.RLock()
	for name := range unames {
		if id, ok := l.cache[name]; ok {
			results[name] = id
		} else {
			missNames = append(missNames, name)
		}
	}
	l.mu.RUnlock()

	if len(missNames) == 0 {
		return results, nil
	}

	parts := make([]string, len(missNames))
	inames := make([]interface{}, len(missNames))

	for i := 0; i != len(missNames); i++ {
		parts[i] = "$" + strconv.Itoa(i+1)
		inames[i] = missNames[i]
	}

	statement := "SELECT timezone_id, name FROM timezones WHERE name IN (" + strings.Join(parts, ",") + ")"

	rows, err := l.db.Query(statement, inames...)

	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for rows.Next() {
		var id int
		var name string

		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}

		results[name] = id
		l.cache[name] = id
	}

	return results, nil
}
