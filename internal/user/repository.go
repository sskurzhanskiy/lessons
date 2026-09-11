package user

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	ByID(ctx context.Context, id int) (User, error)
}
