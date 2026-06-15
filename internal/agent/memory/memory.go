package memory

import "context"

type Store interface {
	Load(ctx context.Context, sessionID string) ([]string, error)
	Append(ctx context.Context, sessionID string, message string) error
}
