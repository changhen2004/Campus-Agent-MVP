package taskapp

import (
	"context"
	"errors"
	"testing"

	taskdomain "campus-agent/internal/domain/task"
)

func TestGetTaskReturnsTaskFromRepository(t *testing.T) {
	t.Parallel()

	repo := &serviceTaskRepo{
		found: taskdomain.Task{
			ID:       1001,
			UserID:   42,
			TaskName: "query_course",
			Status:   taskdomain.StatusSuccess,
			Result:   "done",
		},
	}
	service := NewService(repo, nil)

	task, err := service.GetTask(context.Background(), 1001)
	if err != nil {
		t.Fatalf("GetTask returned error: %v", err)
	}

	if task.ID != 1001 {
		t.Fatalf("id mismatch: got %d want %d", task.ID, 1001)
	}

	if repo.findID != 1001 {
		t.Fatalf("repo id mismatch: got %d want %d", repo.findID, 1001)
	}
}

func TestGetTaskErrorsWhenRepositoryMissing(t *testing.T) {
	t.Parallel()

	service := NewService(nil, nil)
	_, err := service.GetTask(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when repository is not configured")
	}
}

func TestListTasksReturnsTasksFromRepository(t *testing.T) {
	t.Parallel()

	repo := &serviceTaskRepo{
		list: []taskdomain.Task{
			{ID: 2, UserID: 42},
			{ID: 1, UserID: 42},
		},
	}
	service := NewService(repo, nil)

	tasks, err := service.ListTasks(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListTasks returned error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("task count mismatch: got %d want %d", len(tasks), 2)
	}

	if repo.listUserID != 42 {
		t.Fatalf("list user mismatch: got %d want %d", repo.listUserID, 42)
	}
}

type serviceTaskRepo struct {
	findID     int64
	listUserID int64
	found      taskdomain.Task
	list       []taskdomain.Task
	err        error
}

func (r *serviceTaskRepo) Save(_ context.Context, _ taskdomain.Task) error {
	return nil
}

func (r *serviceTaskRepo) UpdateStatus(_ context.Context, _ int64, _ taskdomain.Status, _ string) error {
	return nil
}

func (r *serviceTaskRepo) FindByID(_ context.Context, id int64) (taskdomain.Task, error) {
	r.findID = id
	if r.err != nil {
		return taskdomain.Task{}, r.err
	}
	return r.found, nil
}

func (r *serviceTaskRepo) ListByUser(_ context.Context, userID int64, limit int) ([]taskdomain.Task, error) {
	r.listUserID = userID
	if r.err != nil {
		return nil, r.err
	}
	return r.list, nil
}

var _ error = errors.New("")
