package taskapp

import (
	"context"
	"errors"
	"testing"

	"campus-agent/internal/agent/executor"
	"campus-agent/internal/agent/planner"
	taskdomain "campus-agent/internal/domain/task"
	"campus-agent/internal/mq/message"
)

func TestExecutionHandlerMarksTaskSuccess(t *testing.T) {
	t.Parallel()

	repo := &fakeTaskRepository{}
	executorAgent := &fakeExecutor{
		results: []executor.TaskResult{
			{
				Task:   planner.TaskQueryCourse,
				Status: executor.StatusSuccess,
				Output: "course result",
			},
		},
	}
	handler := NewExecutionHandler(repo, executorAgent)

	err := handler.Handle(context.Background(), message.TaskMessage{
		TaskID:   1001,
		UserID:   42,
		TaskName: string(planner.TaskQueryCourse),
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if len(repo.updates) != 2 {
		t.Fatalf("update count mismatch: got %d want %d", len(repo.updates), 2)
	}

	if repo.updates[0].status != taskdomain.StatusRunning {
		t.Fatalf("first status mismatch: got %q want %q", repo.updates[0].status, taskdomain.StatusRunning)
	}

	if repo.updates[1].status != taskdomain.StatusSuccess {
		t.Fatalf("second status mismatch: got %q want %q", repo.updates[1].status, taskdomain.StatusSuccess)
	}

	if repo.updates[1].result == "" {
		t.Fatal("expected success result to be persisted")
	}

	if len(executorAgent.tasks) != 1 || executorAgent.tasks[0].Name != planner.TaskQueryCourse {
		t.Fatalf("executor tasks mismatch: got %+v", executorAgent.tasks)
	}
}

func TestExecutionHandlerMarksTaskFailed(t *testing.T) {
	t.Parallel()

	repo := &fakeTaskRepository{}
	executorAgent := &fakeExecutor{err: errors.New("tool failed")}
	handler := NewExecutionHandler(repo, executorAgent)

	err := handler.Handle(context.Background(), message.TaskMessage{
		TaskID:   1002,
		UserID:   42,
		TaskName: string(planner.TaskCreateReminder),
	})
	if err == nil {
		t.Fatal("expected Handle to return executor error")
	}

	if len(repo.updates) != 2 {
		t.Fatalf("update count mismatch: got %d want %d", len(repo.updates), 2)
	}

	if repo.updates[0].status != taskdomain.StatusRunning {
		t.Fatalf("first status mismatch: got %q want %q", repo.updates[0].status, taskdomain.StatusRunning)
	}

	if repo.updates[1].status != taskdomain.StatusFailed {
		t.Fatalf("second status mismatch: got %q want %q", repo.updates[1].status, taskdomain.StatusFailed)
	}

	if repo.updates[1].result != "tool failed" {
		t.Fatalf("failure result mismatch: got %q", repo.updates[1].result)
	}
}

type fakeTaskRepository struct {
	updates []taskUpdate
}

type taskUpdate struct {
	id     int64
	status taskdomain.Status
	result string
}

func (r *fakeTaskRepository) Save(_ context.Context, _ taskdomain.Task) error {
	return nil
}

func (r *fakeTaskRepository) UpdateStatus(_ context.Context, id int64, status taskdomain.Status, result string) error {
	r.updates = append(r.updates, taskUpdate{
		id:     id,
		status: status,
		result: result,
	})
	return nil
}

func (r *fakeTaskRepository) FindByID(_ context.Context, id int64) (taskdomain.Task, error) {
	return taskdomain.Task{ID: id}, nil
}

func (r *fakeTaskRepository) ListByUser(_ context.Context, userID int64, limit int) ([]taskdomain.Task, error) {
	return []taskdomain.Task{}, nil
}

type fakeExecutor struct {
	tasks   []planner.Task
	results []executor.TaskResult
	err     error
}

func (e *fakeExecutor) Execute(_ context.Context, tasks []planner.Task) ([]executor.TaskResult, error) {
	e.tasks = append([]planner.Task(nil), tasks...)
	return e.results, e.err
}
