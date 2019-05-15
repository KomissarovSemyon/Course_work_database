package util

import (
	"encoding/json"
	"os"
)

func UnmarshalFromFile(filename string, v interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	return decoder.Decode(v)
}

func MarshalIntoFile(filename string, v interface{}) error {
	f, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	err = enc.Encode(v)
	if err != nil {
		f.Close()
		os.Remove(filename)
		return err
	}

	return nil
}
