package reminderdomain

import (
	"context"
	"time"
)

type Reminder struct {
	ID        int64
	UserID    int64
	Title     string
	Content   string
	TriggerAt time.Time
}

type Repository interface {
	Save(ctx context.Context, reminder Reminder) error
	FindByID(ctx context.Context, id int64) (Reminder, error)
}
