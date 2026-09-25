package user

import "errors"

var (
	ErrInvalidParameter   = errors.New("invalid parameter")
	ErrNotFound           = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string {
	return "validation failed for field: " + e.Field
}
