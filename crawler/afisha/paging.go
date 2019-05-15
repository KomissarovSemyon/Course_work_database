package afisha

import (
	"github.com/pkg/errors"
)

// PagingData contains info for next request paging
type PagingData struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// PagingFunc is callback for PagingLoad
type PagingFunc func(offset int, limit int) (*PagingData, int, error)

// ErrUnexpectedStop is returned if count is zero while there are still results to load
var ErrUnexpectedStop = errors.New("Unexpected Zero count")

// PagingLoad keeps calling eval with adjusted paging params until all results are loaded
func PagingLoad(offset, limit int, eval PagingFunc) error {
	data, cnt, err := eval(offset, limit)
	if err != nil {
		return err
	}

	limit = data.Limit
	offset += cnt

	for offset < data.Total {
		data, cnt, err = eval(offset, limit)
		if err != nil {
			return err
		}
		if cnt == 0 {
			return ErrUnexpectedStop
		}

		limit = data.Limit
		offset += cnt
	}

	return nil
}
