package afisha

import (
	"errors"
	"net/http"
)

// ScheduleCinemaParams holds params for GetScheduleCinema call
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

// ScheduleCinema is result of GetScheduleCinema call
type ScheduleCinema struct {
	Params struct {
		Date Date `json:"date"`
	} `json:"params"`
	Paging PagingData     `json:"paging"`
	Items  []ScheduleItem `json:"items"`
}

// GetScheduleCinema gets cinema schedule (either for cinema or event)
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

// GetScheduleCinemaFull is GetScheduleCinema which loads all results
func GetScheduleCinemaFull(client *http.Client, params *ScheduleCinemaParams) (*ScheduleCinema, error) {
	var result ScheduleCinema
	pagedParams := *params

	err := PagingLoad(params.Offset, params.Limit, func(offset int, limit int) (*PagingData, int, error) {
		pagedParams.Offset = offset
		pagedParams.Limit = limit

		page, err := GetScheduleCinema(client, &pagedParams)
		if err != nil {
			return nil, 0, err
		}

		result.Params.Date = page.Params.Date
		result.Items = append(result.Items, page.Items...)

		return &page.Paging, len(page.Items), nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// RepertoryParams holds params for GetRepetory call
type RepertoryParams struct {
	// If uncpecified, All events in the city would be returned
	PlaceID string `url:"-"`
	City    string `url:"city"`
	Limit   int    `url:"limit,omitempty"`
	Offset  int    `url:"offset,omitempty"`
}

// Repertory is result of GetRepetory call
type Repertory struct {
	Data   []RepertoryItem `json:"data"`
	Paging PagingData      `json:"paging"`
}

// GetRepetory gets reportory of all cinema events in city, or repertory for exact place
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

// GetRepetoryFull is GetRepetory which loads all results
func GetRepetoryFull(client *http.Client, params *RepertoryParams) (*Repertory, error) {
	var result Repertory
	pagedParams := *params

	err := PagingLoad(params.Offset, params.Limit, func(offset int, limit int) (*PagingData, int, error) {
		pagedParams.Offset = offset
		pagedParams.Limit = limit

		page, err := GetRepetory(client, &pagedParams)
		if err != nil {
			return nil, 0, err
		}

		result.Data = append(result.Data, page.Data...)

		return &page.Paging, len(page.Data), nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// PlacesParams holds params for GetPlaces call
type PlacesParams struct {
	City   string `url:"city"`
	Limit  int    `url:"limit,omitempty"`
	Offset int    `url:"offset,omitempty"`
}

// Places is result of GetPlaces call
type Places struct {
	Items  []Place    `json:"items"`
	Paging PagingData `json:"paging"`
}

// GetPlaces gets list of cinemas in city
func GetPlaces(client *http.Client, params *PlacesParams) (*Places, error) {
	endpoint := "/events/cinema/places"
	var resp Places
	err := request(client, endpoint, params, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPlacesFull is GetPlaces which loads all results
func GetPlacesFull(client *http.Client, params *PlacesParams) (*Places, error) {
	var result Places
	pagedParams := *params

	err := PagingLoad(params.Offset, params.Limit, func(offset int, limit int) (*PagingData, int, error) {
		pagedParams.Offset = offset
		pagedParams.Limit = limit

		page, err := GetPlaces(client, &pagedParams)
		if err != nil {
			return nil, 0, err
		}

		result.Items = append(result.Items, page.Items...)

		return &page.Paging, len(page.Items), nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}
