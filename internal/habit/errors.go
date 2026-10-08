package habit

import "errors"

var ErrInvalidParameter = errors.New("invalid parameter")
var ErrNotFound = errors.New("habit not found")
var ErrInvalidCredentials = errors.New("invalid token")
