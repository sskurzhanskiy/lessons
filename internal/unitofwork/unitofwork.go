package unitofwork

import (
	"context"
	"lessonHttp/internal/profile"
	"lessonHttp/internal/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type UnitOfWork struct {
	users    user.Repository
	profiles profile.Repository
}

func NewUnitOfWork(db DBTX) *UnitOfWork {
	return &UnitOfWork{
		users:    user.NewPostgresRepository(db),
		profiles: profile.NewPostgresRepository(db),
	}
}

func New(users user.Repository, profiles profile.Repository) *UnitOfWork {
	return &UnitOfWork{users: users,
		profiles: profiles,
	}
}

func (u *UnitOfWork) Users() user.Repository {
	return u.users
}

func (u *UnitOfWork) Profiles() profile.Repository {
	return u.profiles
}
