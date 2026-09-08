package unitofwork

import (
	"context"
	"lessonHttp/internal/database"
	"lessonHttp/internal/profile"
	"lessonHttp/internal/user"

	"github.com/jackc/pgx/v5"
)

type Manager interface {
	WithTransaction(ctx context.Context, opts pgx.TxOptions, fn func(uow *UnitOfWork) error) error
}

type TransactionManager struct {
	txManager *database.TransactionManager
}

func NewTransactionManager(txManager *database.TransactionManager) *TransactionManager {
	return &TransactionManager{txManager: txManager}
}

func (t *TransactionManager) WithTransaction(ctx context.Context, opts pgx.TxOptions, fn func(uow *UnitOfWork) error) error {
	return t.txManager.WithTransaction(ctx, opts, func(tx pgx.Tx) error {
		uow := New(user.NewPostgresRepository(tx), profile.NewPostgresRepository(tx))
		return fn(uow)
	})
}
