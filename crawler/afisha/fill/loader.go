package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var ErrNotFound = sql.ErrNoRows

type Loader struct {
	cache map[string]int
	mu    sync.RWMutex
	db    *sql.DB
	cfg   LoaderConfig
}

type LoaderConfig struct {
	Table     string
	FieldID   string
	FieldName string
}

func NewLoader(db *sql.DB, config LoaderConfig) *Loader {
	return &Loader{
		cache: make(map[string]int),
		db:    db,
		cfg:   config,
	}
}

func (l *Loader) GetID(name string) (int, error) {
	l.mu.RLock()
	if id, ok := l.cache[name]; ok {
		l.mu.RUnlock()
		return id, nil
	}
	l.mu.RUnlock()

	row := l.db.QueryRow(fmt.Sprintf("SELECT %[2]s FROM %[1]s WHERE %[3]s = $1", l.cfg.Table, l.cfg.FieldID, l.cfg.FieldName), name)

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

func (l *Loader) GetIDs(names []string) (map[string]int, error) {
	results := map[string]int{}

	// Only unique names
	unames := make(map[string]struct{})
	for _, name := range names {
		unames[name] = struct{}{}
	}

	var missNames []string
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

	statement := fmt.Sprintf("SELECT %[2]s, %[3]s FROM %[1]s WHERE %[3]s IN (", l.cfg.Table, l.cfg.FieldID, l.cfg.FieldName) + strings.Join(parts, ",") + ")"

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

type LoadableData interface {
	Fields() []string
	InsertFormat() (string, int)
	Values(map[string]struct{}) []interface{}
	Names() []string
}

func (l *Loader) InsertData(data LoadableData, names map[string]struct{}) (map[string]int, error) {
	values := data.Values(names)
	header := data.Fields()

	if len(values) < 1 {
		return map[string]int{}, nil
	}

	format, fieldCnt := data.InsertFormat()
	format = "( " + format + " )"
	count := len(values) / fieldCnt
	parts := make([]string, count)
	partdata := make([]interface{}, fieldCnt)

	for i := 0; i != count; i++ {
		ii := i * fieldCnt
		for j := 0; j != fieldCnt; j++ {
			partdata[j] = ii + 1 + j
		}

		parts[i] = fmt.Sprintf(format, partdata...)
	}

	statement := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", l.cfg.Table, strings.Join(header, ", ")) +
		strings.Join(parts, ",") +
		fmt.Sprintf(" RETURNING %s, %s", l.cfg.FieldID, l.cfg.FieldName)

	rows, err := l.db.Query(statement, values...)

	if err != nil {
		return nil, err
	}

	results := map[string]int{}
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

func (l *Loader) GetIDsCreating(data LoadableData) (map[string]int, error) {
	names := data.Names()
	result, err := l.GetIDs(names)
	if err != nil {
		return nil, err
	}

	createThese := map[string]struct{}{}
	for _, name := range names {
		if _, ok := result[name]; !ok {
			createThese[name] = struct{}{}
		}
	}

	if len(createThese) == 0 {
		return result, nil
	}

	inserted, err := l.InsertData(data, createThese)
	if err != nil {
		return nil, err
	}

	for name, id := range inserted {
		result[name] = id
	}

	return result, nil
}
