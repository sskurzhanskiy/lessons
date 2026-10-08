package habit

import "time"

type Habit struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	ScheduledAt time.Time `json:"scheduled_at"`
	IsComplete  bool      `json:"is_complete"`
	Comment     string    `json:"comment"`
}

type CreateHabit struct {
	UserID      int
	Title       string
	ScheduledAt time.Time
	IsComplete  bool
	Comment     string
}

type CreateInput struct {
	Title       string    `json:"title"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Comment     string    `json:"comment"`
}

type UpdateInput struct {
	Title       *string    `json:"title"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	IsComplete  *bool      `json:"is_complete,omitempty"`
	Comment     *string    `json:"comment,omitempty"`
}

type UpdateHabit struct {
	Title       *string
	ScheduledAt *time.Time
	IsComplete  *bool
	Comment     *string
}
