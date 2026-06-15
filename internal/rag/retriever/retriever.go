package retriever

import "context"

type Service interface {
	Retrieve(ctx context.Context, query string) ([]string, error)
}
