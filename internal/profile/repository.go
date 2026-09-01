package profile

import "context"

type Repository interface {
	Create(ctx context.Context, profile Profile) error
}
