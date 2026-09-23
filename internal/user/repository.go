package user

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, params CreateUserParams) (User, error)
	ByID(ctx context.Context, id int) (User, error)
	List(ctx context.Context, limit int, offset int) ([]User, error)
}
