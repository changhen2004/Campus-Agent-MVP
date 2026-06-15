package knowledge

import (
	"context"
	"testing"
)

func TestLocalToolSearchesDocuments(t *testing.T) {
	t.Parallel()

	tool := NewLocalTool([]Document{
		{
			ID:      "lab-report",
			Title:   "实验报告提交",
			Content: "实验报告需要通过教务平台提交。",
		},
		{
			ID:      "library",
			Title:   "图书馆开放时间",
			Content: "图书馆开放时间为 8:00 到 22:00。",
		},
	})

	results, err := tool.Search(context.Background(), "实验报告怎么提交")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("result count mismatch: got %d want %d", len(results), 1)
	}

	if results[0] != "实验报告提交: 实验报告需要通过教务平台提交。" {
		t.Fatalf("result mismatch: got %q", results[0])
	}
}
