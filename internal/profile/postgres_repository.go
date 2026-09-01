package profile

import (
	"context"
	"fmt"
	"lessonHttp/internal/database"
)

type PostgresRepository struct {
	db database.DBTX
}

func NewPostgresRepository(db database.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, profile Profile) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO profiles (user_id, bio) VALUES ($1, $2)`,
		profile.UserID,
		profile.Bio,
	)

	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}

	return nil
}
