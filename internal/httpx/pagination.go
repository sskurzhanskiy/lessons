package httpx

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Limit  int
	Offset int
}

const (
	minLimit      = 1
	defaultOffset = 0
	minOffset     = 0
	maxOffset     = -1
)

func ParsePagination(r *http.Request, defaultLimit int, maxLimit int) (Pagination, error) {
	limit, err := parseParameter(r.URL.Query().Get("limit"),
		defaultLimit,
		minLimit,
		maxLimit,
	)
	if err != nil {
		return Pagination{}, err
	}

	offset, err := parseParameter(r.URL.Query().Get("offset"),
		defaultOffset,
		minOffset,
		maxOffset,
	)
	if err != nil {
		return Pagination{}, err
	}

	return Pagination{
		Limit:  limit,
		Offset: offset,
	}, nil
}

func parseParameter(param string, defaultValue int, minValue int, maxValue int) (int, error) {
	if param == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, ErrInvalidPagination
	}

	if value < minValue {
		return 0, ErrInvalidPagination
	}

	if maxValue <= 0 {
		return value, nil
	}

	if value > maxValue {
		return 0, ErrInvalidPagination
	}

	return value, nil
}
