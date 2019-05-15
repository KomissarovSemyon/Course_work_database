package afisha

import (
	"encoding/json"
	"fmt"
	"image/color"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

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
	return json.Marshal(fmt.Sprintf(colorFormat, c.R, c.G, c.B))
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
	return json.Marshal(d.String())
}

// UnmarshalJSON conforms to json.Unmarshaler
func (d *Date) UnmarshalJSON(data []byte) error {
	var strData string
	err := json.Unmarshal(data, &strData)
	if err != nil {
		return err
	}

	if strData == "" {
		*d = Date{}
		return nil
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
