package retriever

import (
	"context"
	"log"

	"campus-agent/pkg/config"

	qdrantrepo "campus-agent/internal/repo/qdrant"
	knowledgetool "campus-agent/internal/tool/knowledge"
)

// Result holds a single retrieval result with its relevance score.
type Result struct {
	Content string
	Score   float64
}

// Retriever decides whether a query relates to the knowledge base
// and returns relevant document content if so.
type Retriever struct {
	qdrant   *qdrantrepo.Retriever
	local    *knowledgetool.LocalTool
	cfg      config.RAGConfig
}

func New(qdrantRetriever *qdrantrepo.Retriever, localTool *knowledgetool.LocalTool, cfg config.RAGConfig) *Retriever {
	return &Retriever{
		qdrant: qdrantRetriever,
		local:  localTool,
		cfg:    cfg,
	}
}

// Retrieve searches for relevant knowledge. Returns the context string and whether it's knowledge-related.
// Uses Qdrant when available, falls back to local keyword search.
func (r *Retriever) Retrieve(ctx context.Context, query string) (context string, isKnowledgeRelated bool) {
	// Try Qdrant first
	if r.qdrant != nil {
		results, err := r.qdrant.Search(ctx, query, 3)
		if err != nil {
			log.Printf("[retriever] qdrant search failed: %v, falling back to local", err)
		} else if len(results) > 0 && results[0].Score >= r.cfg.SimilarityThreshold {
			context = buildContext(results)
			return context, true
		}
		// Below threshold — not knowledge related
		if len(results) > 0 {
			return "", false
		}
	}

	// Fallback: local keyword search
	if r.local != nil {
		docs, err := r.local.Search(ctx, query)
		if err != nil {
			log.Printf("[retriever] local search failed: %v", err)
			return "", false
		}
		if len(docs) > 0 {
			context = buildLocalContext(docs)
			return context, true
		}
	}

	return "", false
}

func buildContext(results []qdrantrepo.SearchResult) string {
	if len(results) == 0 {
		return ""
	}
	parts := make([]string, 0, len(results))
	for _, r := range results {
		if r.Content != "" {
			parts = append(parts, r.Content)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "相关知识库内容：\n" + joinParts(parts)
}

func buildLocalContext(docs []string) string {
	if len(docs) == 0 {
		return ""
	}
	return "相关知识库内容：\n" + joinParts(docs)
}

func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "\n---\n"
		}
		result += p
	}
	return result
}
