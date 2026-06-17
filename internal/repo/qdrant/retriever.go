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
		Limit:          limit * 3, // 过采样以应对块去重
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

		// 向后兼容：如果 doc_id 为空，使用 ID 作为 doc_id
		if sr.DocID == "" {
			sr.DocID = sr.ID
		}

		searchResults = append(searchResults, sr)
	}

	// 合并同一文档的块，保留每个文档得分最高的块
	merged := mergeByDoc(searchResults, limit)
	return merged, nil
}

// mergeByDoc 按 DocID 分组结果，保留每个文档得分最高的块，并拼接相邻块以提供上下文。
func mergeByDoc(results []SearchResult, limit uint64) []SearchResult {
	if len(results) == 0 {
		return nil
	}

	// 按 doc_id 分组
	groups := make(map[string][]SearchResult)
	order := make([]string, 0, len(groups)) // 保留首次出现的顺序

	for _, r := range results {
		if _, exists := groups[r.DocID]; !exists {
			order = append(order, r.DocID)
		}
		groups[r.DocID] = append(groups[r.DocID], r)
	}

	// 对每个文档，按索引排序块并合并内容
	merged := make([]SearchResult, 0, len(groups))
	for _, docID := range order {
		chunks := groups[docID]
		sort.Slice(chunks, func(i, j int) bool {
			return chunks[i].ChunkIndex < chunks[j].ChunkIndex
		})

		// 该文档的最高得分
		bestScore := chunks[0].Score
		bestTitle := chunks[0].Title

		// 拼接块内容（去除重叠尾部）
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

// mergeChunkContent 拼接块内容，去除重叠尾部。
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

		// 查找重叠：prev 的后缀匹配 curr 的前缀
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
