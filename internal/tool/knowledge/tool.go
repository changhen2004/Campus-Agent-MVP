package knowledge

import (
	"context"
	"strings"
	"sync"
)

type Tool interface {
	Search(ctx context.Context, query string) ([]string, error)
}

type Document struct {
	ID      string
	Title   string
	Content string
}

type LocalTool struct {
	mu        sync.RWMutex
	documents []Document
}

func NewLocalTool(documents []Document) *LocalTool {
	copied := make([]Document, len(documents))
	copy(copied, documents)
	return &LocalTool{documents: copied}
}

func (t *LocalTool) Search(_ context.Context, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []string{}, nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	results := make([]string, 0)
	for _, doc := range t.documents {
		if documentMatches(doc, query) {
			results = append(results, doc.Title+": "+doc.Content)
		}
	}
	return results, nil
}

func (t *LocalTool) AddDocuments(documents ...Document) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.documents = append(t.documents, documents...)
}

func documentMatches(doc Document, query string) bool {
	if strings.Contains(doc.Title, query) || strings.Contains(doc.Content, query) {
		return true
	}

	for _, token := range strings.Fields(query) {
		if strings.Contains(doc.Title, token) || strings.Contains(doc.Content, token) {
			return true
		}
	}

	return hasRuneOverlap(query, doc.Title, 2) || hasRuneOverlap(query, doc.Content, 2)
}

func hasRuneOverlap(query string, text string, minLength int) bool {
	runes := []rune(query)
	for start := 0; start < len(runes); start++ {
		for end := start + minLength; end <= len(runes); end++ {
			if strings.Contains(text, string(runes[start:end])) {
				return true
			}
		}
	}
	return false
}

type StubTool struct{}

func NewStubTool() *StubTool {
	return &StubTool{}
}

func (t *StubTool) Search(_ context.Context, query string) ([]string, error) {
	return NewLocalTool([]Document{
		{
			ID:      "lab-report",
			Title:   "实验报告提交",
			Content: "实验报告需要通过教务平台提交。",
		},
	}).Search(context.Background(), query)
}
