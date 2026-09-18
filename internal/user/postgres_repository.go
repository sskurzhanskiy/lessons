package user

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
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, user User) (User, error) {
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (name, age) VALUES ($1, $2) RETURNING id`,
		user.Name,
		user.Age,
	).Scan(&user.ID)

	if err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) ByID(ctx context.Context, id int) (User, error) {
	var user User
	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, age FROM users WHERE id=$1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Age)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}

		return User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) List(ctx context.Context, limit int, offset int) ([]User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, age FROM users ORDER BY id LIMIT $1 OFFSET $2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Age)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}
