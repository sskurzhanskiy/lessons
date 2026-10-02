package habit

import "time"

type Habit struct {
	ID          int
	UserID      int
	Title       string
	ScheduledAt time.Time
	IsComplete  bool
	Comment     string
}

type CreateHabit struct {
	UserID      int
	Title       string
	ScheduledAt time.Time
	Comment     string
}
