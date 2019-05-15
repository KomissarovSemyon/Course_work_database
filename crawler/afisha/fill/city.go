package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

type CityLoader struct {
	cache map[string]int
	mu    sync.RWMutex
	db    *sql.DB
}

func NewCityLoader(db *sql.DB) *CityLoader {
	return &CityLoader{
		cache: make(map[string]int),
		db:    db,
	}
}

func (l *CityLoader) CityID(name string) (int, error) {
	l.mu.RLock()
	if id, ok := l.cache[name]; ok {
		log.Printf("CityID(%s): cached!", name)
		l.mu.RUnlock()
		return id, nil
	}
	l.mu.RUnlock()

	log.Printf("CityID(%s): selecting...", name)
	row := l.db.QueryRow("SELECT city_id FROM cities WHERE ya_name = $1", name)

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

func (l *CityLoader) CityIDs(names []string) (map[string]int, error) {
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

	statement := "SELECT city_id, ya_name FROM cities WHERE ya_name IN (" + strings.Join(parts, ",") + ")"

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

type CityData struct {
	CountryCode [2]byte
	Name        string
	YaName      string
	TimeZoneID  int
}

// XXX: CopyIn?
func (l *CityLoader) InsertCities(cities []CityData) (map[string]int, error) {
	if len(cities) == 0 {
		return map[string]int{}, nil
	}

	parts := make([]string, len(cities))
	idata := make([]interface{}, len(cities)*4)

	fmt.Println("Cities: ", cities)
	for i, city := range cities {
		ii := i * 4
		parts[i] = fmt.Sprintf("( $%d, $%d, $%d, $%d )", ii+1, ii+2, ii+3, ii+4)

		idata[ii+0] = string(city.CountryCode[:])
		idata[ii+1] = city.Name
		idata[ii+2] = city.YaName
		idata[ii+3] = city.TimeZoneID
	}

	statement := "INSERT INTO cities (country_code, name, ya_name, timezone_id) VALUES " +
		strings.Join(parts, ",") +
		" RETURNING city_id, ya_name"

	fmt.Println("Statement: ", statement)
	fmt.Println("Values: ", idata)
	rows, err := l.db.Query(statement, idata...)

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

		l.cache[name] = id
		results[name] = id
	}

	return results, nil
}

func (l *CityLoader) CityIDsCreating(cities []CityData) (map[string]int, error) {
	names := make([]string, len(cities))
	for i, city := range cities {
		names[i] = city.YaName
	}

	result, err := l.CityIDs(names)
	if err != nil {
		return nil, err
	}

	for k, v := range result {
		fmt.Printf("`%v`: `%v`\n", k, v)
	}

	createThese := map[string]*CityData{}
	for i := 0; i != len(cities); i++ {
		city := &cities[i]
		if _, ok := result[city.YaName]; !ok {
			createThese[city.YaName] = city
		}
	}

	if len(createThese) == 0 {
		return result, nil
	}

	insertCities := make([]CityData, len(createThese))
	i := 0
	for _, city := range createThese {
		insertCities[i] = *city
		i++
	}

	inserted, err := l.InsertCities(insertCities)
	if err != nil {
		return nil, err
	}

	for name, id := range inserted {
		result[name] = id
	}

	return result, nil
}
