package mysql

import (
	"context"
	"errors"

	userdomain "campus-agent/internal/domain/user"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) FindByID(_ context.Context, _ int64) (userdomain.User, error) {
	return userdomain.User{}, errors.New("mysql user repository is not implemented")
}
