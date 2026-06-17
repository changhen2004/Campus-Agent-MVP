package retriever

import (
	"context"
	"log"

	"campus-agent/pkg/config"

	qdrantrepo "campus-agent/internal/repo/qdrant"
	knowledgetool "campus-agent/internal/tool/knowledge"
)

// Result 保存单条检索结果及其相关性分数。
type Result struct {
	Content string
	Score   float64
}

// Retriever 判断查询是否与知识库相关，如果相关则返回匹配的文档内容。
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

// Retrieve 检索相关知识。返回上下文字符串以及是否为知识类查询。
// 优先使用 Qdrant 向量检索，不可用时降级为本地关键词搜索。
func (r *Retriever) Retrieve(ctx context.Context, query string) (context string, isKnowledgeRelated bool) {
	// 优先尝试 Qdrant 向量检索
	if r.qdrant != nil {
		results, err := r.qdrant.Search(ctx, query, 3)
		if err != nil {
			log.Printf("[retriever] qdrant 搜索失败: %v，降级为本地搜索", err)
		} else if len(results) > 0 && results[0].Score >= r.cfg.SimilarityThreshold {
			context = buildContext(results)
			return context, true
		}
		// 低于阈值 — 不属于知识类问题
		if len(results) > 0 {
			return "", false
		}
	}

	// 降级方案：本地关键词搜索
	if r.local != nil {
		docs, err := r.local.Search(ctx, query)
		if err != nil {
			log.Printf("[retriever] 本地搜索失败: %v", err)
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
