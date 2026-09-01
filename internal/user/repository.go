package user

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("user not found")

type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	ByID(ctx context.Context, id int) (User, error)
}
