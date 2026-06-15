package vectorstore

import "context"

type Store interface {
	Upsert(ctx context.Context, id string, vector []float32, payload map[string]string) error
	Search(ctx context.Context, vector []float32, topK int) ([]map[string]string, error)
}
