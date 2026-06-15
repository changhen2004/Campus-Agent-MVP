package mysql

import (
	"context"
	"testing"
	"time"

	taskdomain "campus-agent/internal/domain/task"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestTaskRepositorySaveFindAndUpdate(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := NewTaskRepository(db)

	ctx := context.Background()
	task := taskdomain.Task{
		ID:        1001,
		UserID:    42,
		TaskName:  "query_course",
		Status:    taskdomain.StatusPending,
		CreatedAt: time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC),
	}

	if err := repo.Save(ctx, task); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	found, err := repo.FindByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}

	if found.TaskName != task.TaskName {
		t.Fatalf("task name mismatch: got %q want %q", found.TaskName, task.TaskName)
	}

	if err := repo.UpdateStatus(ctx, task.ID, taskdomain.StatusSuccess, "done"); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	updated, err := repo.FindByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("FindByID after update returned error: %v", err)
	}

	if updated.Status != taskdomain.StatusSuccess {
		t.Fatalf("status mismatch: got %q want %q", updated.Status, taskdomain.StatusSuccess)
	}

	if updated.Result != "done" {
		t.Fatalf("result mismatch: got %q want %q", updated.Result, "done")
	}
}

func TestTaskRepositoryListByUser(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)
	repo := NewTaskRepository(db)

	ctx := context.Background()
	tasks := []taskdomain.Task{
		{
			ID:        1,
			UserID:    42,
			TaskName:  "query_course",
			Status:    taskdomain.StatusPending,
			CreatedAt: time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			UserID:    42,
			TaskName:  "create_reminder",
			Status:    taskdomain.StatusSuccess,
			CreatedAt: time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:        3,
			UserID:    7,
			TaskName:  "search_knowledge",
			Status:    taskdomain.StatusSuccess,
			CreatedAt: time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
		},
	}

	for _, task := range tasks {
		if err := repo.Save(ctx, task); err != nil {
			t.Fatalf("Save returned error: %v", err)
		}
	}

	found, err := repo.ListByUser(ctx, 42, 10)
	if err != nil {
		t.Fatalf("ListByUser returned error: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("task count mismatch: got %d want %d", len(found), 2)
	}

	if found[0].ID != 2 || found[1].ID != 1 {
		t.Fatalf("order mismatch: got ids %d, %d", found[0].ID, found[1].ID)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}
