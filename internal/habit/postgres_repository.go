package habit

import (
	"context"
	"fmt"
	"lessonHttp/internal/database"
)

type PostgresRepository struct {
	db database.DBTX
}

func NewPostgresRepository(db database.DBTX) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(ctx context.Context, param CreateHabit) (Habit, error) {
	var habit Habit
	err := r.db.QueryRow(ctx,
		`INSERT INTO habits (user_id, title, scheduled_at, comment) VALUES($1, $2, $3, $4) RETURNING id, user_id, title, scheduled_at, is_complete, comment`,
		param.UserID,
		param.Title,
		param.ScheduledAt,
		param.Comment,
	).Scan(&habit.ID,
		&habit.UserID,
		&habit.Title,
		&habit.ScheduledAt,
		&habit.IsComplete,
		&habit.Comment)
	if err != nil {
		return Habit{}, fmt.Errorf("create habit: %w", err)
	}

	return habit, nil
}
