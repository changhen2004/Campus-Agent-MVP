package qdrant

import (
	"context"
	"fmt"
	"sort"

	"campus-agent/internal/ai/embedder"

	pb "github.com/qdrant/go-client/qdrant"
)

type Retriever struct {
	client   *Client
	embedder *embedder.Embedder
}

type SearchResult struct {
	ID         string
	DocID      string
	Title      string
	Content    string
	ChunkIndex int
	Score      float64
}

func NewRetriever(client *Client, embedder *embedder.Embedder) *Retriever {
	return &Retriever{client: client, embedder: embedder}
}

func (r *Retriever) Search(ctx context.Context, query string, limit uint64) ([]SearchResult, error) {
	vector64, err := r.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	vector32 := make([]float32, len(vector64))
	for i, v := range vector64 {
		vector32[i] = float32(v)
	}

	results, err := r.client.points.Search(ctx, &pb.SearchPoints{
		CollectionName: r.client.collection,
		Vector:         vector32,
		Limit:          limit * 3, // oversample to account for chunk dedup
		WithPayload:    pb.NewWithPayloadEnable(true),
	})
	if err != nil {
		return nil, fmt.Errorf("search qdrant: %w", err)
	}

	searchResults := make([]SearchResult, 0, len(results.GetResult()))
	for _, point := range results.GetResult() {
		payload := point.GetPayload()
		sr := SearchResult{Score: float64(point.GetScore())}

		if v, ok := payload["title"]; ok {
			sr.Title = v.GetStringValue()
		}
		if v, ok := payload["content"]; ok {
			sr.Content = v.GetStringValue()
		}
		if v, ok := payload["doc_id"]; ok {
			sr.DocID = v.GetStringValue()
		}
		if v, ok := payload["chunk_index"]; ok {
			sr.ChunkIndex = int(v.GetIntegerValue())
		}
		if pid := point.GetId(); pid != nil {
			sr.ID = pid.GetUuid()
		}

		// Backward compat: if doc_id is empty, use ID as doc_id
		if sr.DocID == "" {
			sr.DocID = sr.ID
		}

		searchResults = append(searchResults, sr)
	}

	// Merge chunks from the same document, keep top-scoring chunk per doc
	merged := mergeByDoc(searchResults, limit)
	return merged, nil
}

// mergeByDoc groups results by DocID, keeps the best-scoring chunk from each
// document, and concatenates adjacent chunks for context.
func mergeByDoc(results []SearchResult, limit uint64) []SearchResult {
	if len(results) == 0 {
		return nil
	}

	// Group by doc_id
	groups := make(map[string][]SearchResult)
	order := make([]string, 0, len(groups)) // preserve first-seen order

	for _, r := range results {
		if _, exists := groups[r.DocID]; !exists {
			order = append(order, r.DocID)
		}
		groups[r.DocID] = append(groups[r.DocID], r)
	}

	// For each doc, sort chunks by index and merge content
	merged := make([]SearchResult, 0, len(groups))
	for _, docID := range order {
		chunks := groups[docID]
		sort.Slice(chunks, func(i, j int) bool {
			return chunks[i].ChunkIndex < chunks[j].ChunkIndex
		})

		// Best score from this doc
		bestScore := chunks[0].Score
		bestTitle := chunks[0].Title

		// Concatenate chunk content (dedup overlapping tails)
		content := mergeChunkContent(chunks)

		merged = append(merged, SearchResult{
			ID:      docID,
			DocID:   docID,
			Title:   bestTitle,
			Content: content,
			Score:   bestScore,
		})
	}

	if uint64(len(merged)) > limit {
		merged = merged[:limit]
	}
	return merged
}

// mergeChunkContent concatenates chunks, removing overlap tails.
func mergeChunkContent(chunks []SearchResult) string {
	if len(chunks) == 0 {
		return ""
	}
	if len(chunks) == 1 {
		return chunks[0].Content
	}

	result := chunks[0].Content
	for i := 1; i < len(chunks); i++ {
		prev := []rune(chunks[i-1].Content)
		curr := []rune(chunks[i].Content)

		// Find overlap: suffix of prev that matches prefix of curr
		overlapLen := 0
		maxOverlap := len(prev)
		if len(curr) < maxOverlap {
			maxOverlap = len(curr)
		}
		for o := maxOverlap; o > 10; o-- {
			if runesEqual(prev[len(prev)-o:], curr[:o]) {
				overlapLen = o
				break
			}
		}

		if overlapLen > 0 {
			result += string(curr[overlapLen:])
		} else {
			result += string(curr)
		}
	}
	return result
}

func runesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
