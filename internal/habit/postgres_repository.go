package habit

import (
	"context"
	"errors"
	"fmt"
	"lessonHttp/internal/database"

	"github.com/jackc/pgx/v5"
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

func (r *PostgresRepository) List(ctx context.Context, userID int, limit int, offset int) ([]Habit, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, title, scheduled_at, is_complete, comment FROM habits WHERE user_id=$1 ORDER BY id LIMIT $2 OFFSET $3`,
		userID,
		limit,
		offset)
	if err != nil {
		return nil, fmt.Errorf("get list habits %w", err)
	}
	defer rows.Close()

	habits := make([]Habit, 0)
	for rows.Next() {
		var habit Habit
		err := rows.Scan(&habit.ID, &habit.UserID, &habit.Title, &habit.ScheduledAt, &habit.IsComplete, &habit.Comment)
		if err != nil {
			return nil, fmt.Errorf("scan failed %w", err)
		}
		habits = append(habits, habit)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row interation error %w", err)
	}

	return habits, nil
}

func (r *PostgresRepository) ByID(ctx context.Context, userID int, habitID int) (Habit, error) {
	var habit Habit
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, title, scheduled_at, is_complete, comment FROM habits WHERE id=$1 AND user_id=$2`,
		habitID,
		userID,
	).Scan(&habit.ID, &habit.UserID, &habit.Title, &habit.ScheduledAt, &habit.IsComplete, &habit.Comment)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Habit{}, ErrNotFound
		}
		return Habit{}, fmt.Errorf("get habit by id: %w", err)
	}

	return habit, nil
}

func (r *PostgresRepository) Update(ctx context.Context, userID int, habitID int, update UpdateHabit) (Habit, error) {
	var habit Habit
	err := r.db.QueryRow(ctx,
		`UPDATE habits SET title=COALESCE($1, title), scheduled_at=COALESCE($2::timestamptz, scheduled_at), is_complete=COALESCE($3, is_complete), comment=COALESCE($4, comment) WHERE id=$5 AND user_id=$6 RETURNING id, user_id, title, scheduled_at, is_complete, comment`,
		update.Title,
		update.ScheduledAt,
		update.IsComplete,
		update.Comment,
		habitID,
		userID,
	).Scan(&habit.ID, &habit.UserID, &habit.Title, &habit.ScheduledAt, &habit.IsComplete, &habit.Comment)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Habit{}, ErrNotFound
		}
		return Habit{}, fmt.Errorf("update habit %w", err)
	}

	return habit, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, userID int, habitID int) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM habits WHERE id=$1 AND user_id=$2`,
		habitID,
		userID)
	if err != nil {
		return fmt.Errorf("delete habit %d %w", habitID, err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
