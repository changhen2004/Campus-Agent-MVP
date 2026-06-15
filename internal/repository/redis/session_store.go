package redis

import "context"

type SessionStore struct{}

func NewSessionStore() *SessionStore {
	return &SessionStore{}
}

func (s *SessionStore) Load(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func (s *SessionStore) Append(_ context.Context, _ string, _ string) error {
	return nil
}
