package habit

import "context"

type Repository interface {
	Create(ctx context.Context, param CreateHabit) (Habit, error)
}
