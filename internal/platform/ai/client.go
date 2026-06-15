package ai

import "context"

type ChatClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

type EmbeddingClient interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}
