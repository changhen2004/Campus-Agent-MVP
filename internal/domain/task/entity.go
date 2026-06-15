package taskdomain

import (
	"context"
	"time"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

type Task struct {
	ID        int64
	UserID    int64
	TaskName  string
	Status    Status
	Result    string
	CreatedAt time.Time
}

type Repository interface {
	Save(ctx context.Context, task Task) error
	UpdateStatus(ctx context.Context, id int64, status Status, result string) error
	FindByID(ctx context.Context, id int64) (Task, error)
	ListByUser(ctx context.Context, userID int64, limit int) ([]Task, error)
}
