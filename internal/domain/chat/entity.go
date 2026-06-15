package chatdomain

import (
	"context"
	"time"
)

type Message struct {
	ID        int64
	UserID    int64
	Role      string
	Content   string
	CreatedAt time.Time
}

type Repository interface {
	Save(ctx context.Context, message Message) error
	ListByUser(ctx context.Context, userID int64, limit int) ([]Message, error)
}
