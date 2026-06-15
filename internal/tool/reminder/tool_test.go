package reminder

import (
	"context"
	"testing"

	reminderdomain "campus-agent/internal/domain/reminder"
)

func TestRepositoryToolCreatesReminder(t *testing.T) {
	t.Parallel()

	repo := &fakeReminderRepository{}
	tool := NewRepositoryTool(repo)

	reminder := reminderdomain.Reminder{
		UserID:  42,
		Title:   "数据库实验",
		Content: "提醒我完成数据库实验",
	}

	if err := tool.Create(context.Background(), reminder); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("saved count mismatch: got %d want %d", len(repo.saved), 1)
	}

	if repo.saved[0].Title != reminder.Title {
		t.Fatalf("title mismatch: got %q want %q", repo.saved[0].Title, reminder.Title)
	}
}

type fakeReminderRepository struct {
	saved []reminderdomain.Reminder
}

func (r *fakeReminderRepository) Save(_ context.Context, reminder reminderdomain.Reminder) error {
	r.saved = append(r.saved, reminder)
	return nil
}

func (r *fakeReminderRepository) FindByID(_ context.Context, id int64) (reminderdomain.Reminder, error) {
	return reminderdomain.Reminder{ID: id}, nil
}
