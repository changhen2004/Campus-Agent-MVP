package user

import (
	"context"

	userdomain "campus-agent/internal/domain/user"
)

type Tool interface {
	GetProfile(ctx context.Context, userID int64) (userdomain.User, error)
}

type StubTool struct{}

func NewStubTool() *StubTool {
	return &StubTool{}
}

func (t *StubTool) GetProfile(_ context.Context, userID int64) (userdomain.User, error) {
	return userdomain.User{
		ID:       userID,
		Username: "demo-user",
		Email:    "demo@example.com",
		Role:     userdomain.RoleUser,
	}, nil
}
