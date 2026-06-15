package mysql

import (
	"context"
	"testing"
	"time"

	reminderdomain "campus-agent/internal/domain/reminder"
)

func TestReminderRepositorySaveAndFindByID(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := NewReminderRepository(db)

	ctx := context.Background()
	reminder := reminderdomain.Reminder{
		ID:        2001,
		UserID:    42,
		Title:     "数据库实验",
		Content:   "提醒我完成数据库实验",
		TriggerAt: time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC),
	}

	if err := repo.Save(ctx, reminder); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	found, err := repo.FindByID(ctx, reminder.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}

	if found.UserID != reminder.UserID {
		t.Fatalf("user mismatch: got %d want %d", found.UserID, reminder.UserID)
	}

	if found.Title != reminder.Title {
		t.Fatalf("title mismatch: got %q want %q", found.Title, reminder.Title)
	}
}
