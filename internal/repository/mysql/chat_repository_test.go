package mysql

import (
	"context"
	"testing"
	"time"

	chatdomain "campus-agent/internal/domain/chat"
)

func TestChatRepositorySaveAndListByUser(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := NewChatRepository(db)

	ctx := context.Background()
	messages := []chatdomain.Message{
		{
			ID:        1,
			UserID:    7,
			Role:      "user",
			Content:   "first",
			CreatedAt: time.Date(2026, 6, 14, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			UserID:    7,
			Role:      "assistant",
			Content:   "second",
			CreatedAt: time.Date(2026, 6, 14, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:        3,
			UserID:    8,
			Role:      "user",
			Content:   "other-user",
			CreatedAt: time.Date(2026, 6, 14, 11, 0, 0, 0, time.UTC),
		},
	}

	for _, msg := range messages {
		if err := repo.Save(ctx, msg); err != nil {
			t.Fatalf("Save returned error: %v", err)
		}
	}

	found, err := repo.ListByUser(ctx, 7, 1)
	if err != nil {
		t.Fatalf("ListByUser returned error: %v", err)
	}

	if len(found) != 1 {
		t.Fatalf("message count mismatch: got %d want %d", len(found), 1)
	}

	if found[0].Content != "second" {
		t.Fatalf("latest message mismatch: got %q want %q", found[0].Content, "second")
	}
}
