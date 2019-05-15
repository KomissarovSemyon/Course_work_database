package afisha

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

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

	log.Printf("INFO: Fetching URL: `%v`", u)
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
