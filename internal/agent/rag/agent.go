package rag

import (
	"context"
	"strings"
)

type Retriever interface {
	Retrieve(ctx context.Context, query string) ([]string, error)
}

type Agent struct {
	retriever Retriever
}

func NewAgent(retriever Retriever) *Agent {
	return &Agent{retriever: retriever}
}

func (a *Agent) Answer(ctx context.Context, query string) (string, error) {
	if a.retriever == nil {
		return "rag retriever is not configured", nil
	}

	documents, err := a.retriever.Retrieve(ctx, query)
	if err != nil {
		return "", err
	}
	if len(documents) == 0 {
		return "no knowledge found", nil
	}
	return strings.Join(documents, "\n"), nil
}
