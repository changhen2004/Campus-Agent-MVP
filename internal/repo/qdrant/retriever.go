package qdrant

import (
	"context"
	"fmt"

	"campus-agent/internal/ai/embedder"

	pb "github.com/qdrant/go-client/qdrant"
)

type Retriever struct {
	client   *Client
	embedder *embedder.Embedder
}

type SearchResult struct {
	ID      string
	Title   string
	Content string
	Score   float64
}

func NewRetriever(client *Client, embedder *embedder.Embedder) *Retriever {
	return &Retriever{client: client, embedder: embedder}
}

func (r *Retriever) Search(ctx context.Context, query string, limit uint64) ([]SearchResult, error) {
	vector64, err := r.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	// Convert float64 to float32 for Qdrant
	vector32 := make([]float32, len(vector64))
	for i, v := range vector64 {
		vector32[i] = float32(v)
	}

	results, err := r.client.points.Search(ctx, &pb.SearchPoints{
		CollectionName: r.client.collection,
		Vector:         vector32,
		Limit:          limit,
		WithPayload:    pb.NewWithPayloadEnable(true),
	})
	if err != nil {
		return nil, fmt.Errorf("search qdrant: %w", err)
	}

	searchResults := make([]SearchResult, 0, len(results.GetResult()))
	for _, point := range results.GetResult() {
		payload := point.GetPayload()
		title := ""
		content := ""
		if v, ok := payload["title"]; ok {
			title = v.GetStringValue()
		}
		if v, ok := payload["content"]; ok {
			content = v.GetStringValue()
		}

		var id string
		if pid := point.GetId(); pid != nil {
			id = pid.GetUuid()
		}

		searchResults = append(searchResults, SearchResult{
			ID:      id,
			Title:   title,
			Content: content,
			Score:   float64(point.GetScore()),
		})
	}

	return searchResults, nil
}
