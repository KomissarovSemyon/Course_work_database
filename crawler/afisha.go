package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

func unmarshalFromFile(filename string, v interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	return decoder.Decode(v)
}

func loadYaCities() ([]string, error) {
	const yaCitiesFilename = "ya-cities.json"
	var cities []string
	if err := unmarshalFromFile(yaCitiesFilename, &cities); err != nil {
		return nil, err
	}

	return cities, nil
}

type KinopoiskScoreData struct {
	URL   string  `json:"url"`
	Value float32 `json:"value"`
	Votes int     `json:"votes"`
}

// Event is event data in Yandex.Afisha API
type Event struct {
	ID            string             `json:"id"`
	URL           string             `json:"url"`
	Title         string             `json:"title"`
	OriginalTitle string             `json:"originalTitle"`
	Kinopoisk     KinopoiskScoreData `json:"kinopoisk"`
}

// Only full color format is supported
const colorFormat = "#%02x%02x%02x"

type Color color.RGBA

// MarshalJSON conforms to json.Marshaler
func (c Color) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(colorFormat, c.R, c.G, c.B)), nil
}

// UnmarshalJSON conforms to json.Unmarshaler
func (c *Color) UnmarshalJSON(data []byte) error {
	var strData string
	err := json.Unmarshal(data, &strData)
	if err != nil {
		return err
	}

	n, err := fmt.Sscanf(strData, colorFormat, &c.R, &c.G, &c.B)

	if err == nil && n != 3 {
		err = errors.Errorf("Expected to scan 3 items, but got %d instead", n)
	}
	if err != nil {
		return errors.Wrapf(err, "While parsing %v as color", strData)
	}

	c.A = 255
	return nil
}

// YYYY-MM-DD
const dateLayout = "2006-01-02"

// Date is date in YYYY-MM-DD format used in various parts of the API
type Date time.Time

// ParseDate parses Date from string
func ParseDate(value string) (Date, error) {
	t, err := time.Parse(dateLayout, value)
	return Date(t), err
}

func (d Date) String() string {
	return time.Time(d).Format(dateLayout)
}

// MarshalJSON conforms to json.Marshaler
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalJSON conforms to json.Unmarshaler
func (d *Date) UnmarshalJSON(data []byte) error {
	var strData string
	err := json.Unmarshal(data, &strData)
	if err != nil {
		return err
	}

	parsed, err := ParseDate(strData)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

// EncodeValues conforms to query.Encoder
func (d Date) EncodeValues(key string, v *url.Values) error {
	v.Add(key, d.String())
	return nil
}

// ScheduleInfo is event schedule summary included in RepertoryItem
type ScheduleInfo struct {
	Dates        []Date `json:"dates"`
	DateStarted  Date   `json:"dateStarted"`
	DateEnd      Date   `json:"dateEnd"`
	DateReleased Date   `json:"dateReleased"`
	PlacesTotal  int    `json:"placedTotal"`
}

// RepertoryItem is an item in all events list
type RepertoryItem struct {
	Event        Event        `json:"event"`
	ScheduleInfo ScheduleInfo `json:"scheduleInfo"`
}

// City -- godoc fuck off
type City struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	GeoID    int    `json:"geoid"`
	TimeZone string `json:"timezone"`
}

// MetroInfo holds info about metro station
type MetroInfo struct {
	Name   string   `json:"name"`
	Colors []string `json:"colors"`
}

type Coordinates struct {
	Longitude float32 `json:"longitude"`
	Latitude  float32 `json:"latitude"`
}

// Place is a speicific place (i.e. cinema) to which schedule is applied
type Place struct {
	ID          string      `json:"id"`
	URL         string      `json:"url"`
	Title       string      `json:"title"`
	Address     string      `json:"address"`
	City        City        `json:"city"`
	Metro       []MetroInfo `json:"metro"`
	Coordinates Coordinates `json:"coordinates"`
	Links       []string    `json:"links"`

	LogoColor Color `json:"logoColor"`
	BGColor   Color `json:"bgColor"`
}

// NamedItem is helper type which unwraps {"name": "str"} to "str"
type NamedItem string

type namedItemStruct struct {
	Name string `json:"name"`
}

// UnmarshalJSON conforms to json.Unmarshaler interface
func (it *NamedItem) UnmarshalJSON(data []byte) error {
	var named namedItemStruct
	if err := json.Unmarshal(data, &named); err != nil {
		return err
	}
	*it = NamedItem(named.Name)
	return nil
}

// MarshalJSON conforms to json.Marshaler interface
func (it NamedItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(namedItemStruct{
		Name: string(it),
	})
}

// PriceRange is ticket price range with currency
type PriceRange struct {
	Currency string `json:"currency"`
	Min      int    `json:"min"`
	Max      int    `json:"max"`
}

// TicketInfo holds info on specific "ticket" options
type TicketInfo struct {
	ID    string     `json:"id"`
	Price PriceRange `json:"price"`
}

// ScheduleSession holds info about exact session in schedule
type ScheduleSession struct {
	Date     Date       `json:"date"`
	Datetime string     `json:"datetime"`
	Ticket   TicketInfo `json:"ticket"`
	HallName string     `json:"hall"`
}

// ScheduleSubItem is an item with specific type of sessions in it
type ScheduleSubItem struct {
	Format   NamedItem         `json:"format"`
	Tags     []NamedItem       `json:"tags"`
	Sessions []ScheduleSession `json:"sessions"`
}

// ScheduleItem is an item in movie schedule with place and subitems
type ScheduleItem struct {
	Date     Date              `json:"date"`
	Place    *Place            `json:"place"`
	Event    *Event            `json:"event"`
	Schedule []ScheduleSubItem `json:"schedule"`
}

type PagingData struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type ScheduleCinema struct {
	Params struct {
		Date Date `json:"date"`
	} `json:"params"`
	Paging PagingData     `json:"paging"`
	Items  []ScheduleItem `json:"items"`
}

const (
	apiBaseURL = "https://afisha.yandex.ru/api/"
)

func request(client *http.Client, endpoint string, params interface{}, resp interface{}) error {
	u, err := url.Parse(apiBaseURL + endpoint)
	if err != nil {
		return err
	}

	q, err := query.Values(params)
	if err != nil {
		return err
	}
	u.RawQuery = q.Encode()

	fmt.Println("URL: ", u.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&resp); err != nil {
		return err
	}

	return nil
}

type ScheduleCinemaParams struct {
	EventID string `url:"-"`
	PlaceID string `url:"-"`
	Date    Date   `url:"date"`
	City    string `url:"city,omitempty"`
	Limit   int    `url:"limit,omitempty"`
	Offset  int    `url:"offset,omitempty"`
}

type scheduleCinemaResponse struct {
	Schedule ScheduleCinema `json:"schedule"`
}

func GetScheduleCinema(client *http.Client, params *ScheduleCinemaParams) (*ScheduleCinema, error) {
	var endpoint string

	switch {
	case params.EventID != "" && params.PlaceID != "":
		return nil, errors.New("Set either EventID or PlaceID, not both")
	case params.EventID != "":
		endpoint = "events/" + params.EventID + "/schedule_cinema"
	case params.PlaceID != "":
		endpoint = "places/" + params.PlaceID + "/schedule_cinema"
	default:
		return nil, errors.New("Either EventID or PlaceID is required")
	}

	var resp scheduleCinemaResponse
	err := request(client, endpoint, params, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Schedule, nil
}

type RepertoryParams struct {
	// If uncpecified, All events in the city would be returned
	PlaceID string `url:"-"`
	City    string `url:"city"`
	Limit   int    `url:"limit,omitempty"`
	Offset  int    `url:"offset,omitempty"`
}

type Repertory struct {
	Data   []RepertoryItem `json:"data"`
	Paging PagingData      `json:"paging"`
}

func GetRepetory(client *http.Client, params *RepertoryParams) (*Repertory, error) {
	var endpoint string

	if params.PlaceID == "" {
		endpoint = "events/selection/all-events-cinema"
	} else {
		endpoint = "places/" + params.PlaceID + "/repertory"
	}

	var resp Repertory
	err := request(client, endpoint, params, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type PlacesParams struct {
	City   string `url:"city"`
	Limit  int    `url:"limit,omitempty"`
	Offset int    `url:"offset,omitempty"`
}

type Places struct {
	Items  []Place    `json:"items"`
	Paging PagingData `json:"paging"`
}

func GetPlaces(client *http.Client, params *PlacesParams) (*Places, error) {
	endpoint := "/events/cinema/places"
	var resp Places
	err := request(client, endpoint, params, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func main() {
	// fmt.Println("hi")

	// cities, err := loadYaCities()
	// if err != nil {
	// 	panic(err)
	// }

	// for _, city := range cities {
	// 	fmt.Println(city)
	// }

	// {
	// 	var schedule scheduleCinemaResponse
	// 	err = unmarshalFromFile("schedule_cinema.json", &schedule)
	// 	if err != nil {
	// 		panic(fmt.Sprint("Failed to unmarshal schedule_cinema: ", err))
	// 	}
	// }

	// {
	// 	params := ScheduleCinemaParams{
	// 		EventID: "5a499676a03db3258b371951",
	// 		Limit:   2,
	// 		Offset:  0,
	// 		City:    "moscow",
	// 	}
	// 	params.Date, _ = ParseDate("2019-05-17")

	// 	fmt.Println("Getting schedule with params:")
	// 	pp.Print(params)

	// 	schedule, err := GetScheduleCinema(http.DefaultClient, &params)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	pp.Print(*schedule)
	// }

	// {
	// 	params := ScheduleCinemaParams{
	// 		PlaceID: "584f9a06ca15c798b73e8b9f",
	// 		Limit:   2,
	// 		Offset:  0,
	// 		City:    "moscow",
	// 	}
	// 	params.Date, _ = ParseDate("2019-05-17")

	// 	fmt.Println("Getting schedule with params:")
	// 	pp.Print(params)

	// 	schedule, err := GetScheduleCinema(http.DefaultClient, &params)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	pp.Print(*schedule)
	// }

	// {
	// 	params := RepertoryParams{
	// 		Limit:  3,
	// 		Offset: 0,
	// 		City:   "saint-petersburg",
	// 	}

	// 	allEvents, err := GetRepetory(http.DefaultClient, &params)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	pp.Print(allEvents)
	// }

	{
		params := PlacesParams{
			Limit: 1,
			City:  "stavropol",
		}

		places, err := GetPlaces(http.DefaultClient, &params)
		if err != nil {
			panic(err)
		}
		pp.Print(places)
	}
}
