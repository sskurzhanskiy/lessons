package user

import "errors"

var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrNotFound         = errors.New("user not found")
)

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string {
	return "validation failed for field: " + e.Field
}
