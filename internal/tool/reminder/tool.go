package reminder

import (
	"context"

	reminderdomain "campus-agent/internal/domain/reminder"
)

type Tool interface {
	Create(ctx context.Context, reminder reminderdomain.Reminder) error
}

type RepositoryTool struct {
	repo reminderdomain.Repository
}

func NewRepositoryTool(repo reminderdomain.Repository) *RepositoryTool {
	return &RepositoryTool{repo: repo}
}

func (t *RepositoryTool) Create(ctx context.Context, reminder reminderdomain.Reminder) error {
	return t.repo.Save(ctx, reminder)
}

type StubTool struct{}

func NewStubTool() *StubTool {
	return &StubTool{}
}

func (t *StubTool) Create(_ context.Context, _ reminderdomain.Reminder) error {
	return nil
}
