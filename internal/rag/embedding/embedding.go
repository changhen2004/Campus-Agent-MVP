package embedding

import "context"

type Service interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}
