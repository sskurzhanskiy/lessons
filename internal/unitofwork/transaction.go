package unitofwork

import (
	"context"
	"lessonHttp/internal/database"

	"github.com/jackc/pgx/v5"
)

type TransactionManager struct {
	txManager *database.TransactionManager
}

func NewTransactionManager(txManager *database.TransactionManager) *TransactionManager {
	return &TransactionManager{txManager: txManager}
}

func (t *TransactionManager) WithTransaction(ctx context.Context, fn func(uow *UnitOfWork) error) error {
	return t.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		uow := NewUnitOfWork(tx)
		return fn(uow)
	})
}
