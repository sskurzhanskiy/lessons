package habit

import "context"

type Repository interface {
	Create(ctx context.Context, param CreateHabit) (Habit, error)
	List(ctx context.Context, userID int, limit int, offset int) ([]Habit, error)
	ByID(ctx context.Context, userID int, habitID int) (Habit, error)
	Update(ctx context.Context, userID int, habitID int, update UpdateHabit) (Habit, error)
	Delete(ctx context.Context, userID int, habitID int) error
}
