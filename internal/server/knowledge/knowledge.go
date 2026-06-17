package knowledge

import (
	"context"
	"fmt"
	"log"

	localknowledge "campus-agent/internal/knowledge/local"
	"campus-agent/internal/repo/qdrant"
	knowledgetool "campus-agent/internal/tool/knowledge"
)

type Service struct {
	localStore *knowledgetool.LocalTool
	qdrantIdx  *qdrant.Indexer
}

func NewService(localStore *knowledgetool.LocalTool, qdrantIndexer *qdrant.Indexer) *Service {
	return &Service{
		localStore: localStore,
		qdrantIdx:  qdrantIndexer,
	}
}

func (s *Service) Upload(ctx context.Context, filename string, content []byte) error {
	doc, err := localknowledge.ParseUpload(filename, content)
	if err != nil {
		return fmt.Errorf("parse uploaded file: %w", err)
	}

	// 添加到本地存储作为降级方案
	s.localStore.AddDocuments(doc)

	// 索引到 Qdrant（如果可用）
	if s.qdrantIdx != nil {
		if err := s.qdrantIdx.IndexDocuments(ctx, []qdrant.Document{
			{ID: doc.ID, Title: doc.Title, Content: doc.Content},
		}); err != nil {
			log.Printf("[knowledge] qdrant index failed: %v (still available in local store)", err)
		}
	}

	return nil
}

// LoadDocs 从知识库目录加载 Markdown 文档。
func (s *Service) LoadDocs(docs []knowledgetool.Document) {
	s.localStore.AddDocuments(docs...)
}

// IndexDocs 将文档索引到 Qdrant。
func (s *Service) IndexDocs(ctx context.Context, docs []knowledgetool.Document) error {
	if s.qdrantIdx == nil {
		return nil
	}

	qdocs := make([]qdrant.Document, 0, len(docs))
	for _, d := range docs {
		qdocs = append(qdocs, qdrant.Document{
			ID:      d.ID,
			Title:   d.Title,
			Content: d.Content,
		})
	}
	return s.qdrantIdx.IndexDocuments(ctx, qdocs)
}
