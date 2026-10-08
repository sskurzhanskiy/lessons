package httpx

import (
	"errors"
)

var ErrInternalServer = errors.New("internal server error")
var ErrInvalidPagination = errors.New("invalid pagination")
