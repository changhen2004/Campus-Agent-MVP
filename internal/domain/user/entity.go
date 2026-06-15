package userdomain

import (
	"context"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID        int64
	Username  string
	Email     string
	Role      Role
	CreatedAt time.Time
}

type Repository interface {
	FindByID(ctx context.Context, id int64) (User, error)
}
