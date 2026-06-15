package router

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"campus-agent/internal/agent/executor"
	"campus-agent/internal/agent/planner"
	"campus-agent/internal/api/handler"
	chatapp "campus-agent/internal/app/chat"
	taskapp "campus-agent/internal/app/task"
	taskdomain "campus-agent/internal/domain/task"
	"campus-agent/internal/mq/producer"
	"github.com/gin-gonic/gin"
)

func TestHealthz(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	engine := New(
		handler.NewChatHandler(chatServiceStub()),
		handler.NewTaskHandler(taskServiceStub()),
	)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d", rec.Code, http.StatusOK)
	}

	if body := rec.Body.String(); body != "ok" {
		t.Fatalf("body mismatch: got %q want %q", body, "ok")
	}
}

func TestChatRoute(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	engine := New(
		handler.NewChatHandler(chatServiceStub()),
		handler.NewTaskHandler(taskServiceStub()),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(`{"user_id":1,"message":"帮我查询明天课程"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestGetTaskRoute(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	engine := New(
		handler.NewChatHandler(chatServiceStub()),
		handler.NewTaskHandler(taskServiceStub()),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/1001", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestGetTaskRouteRejectsInvalidID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	engine := New(
		handler.NewChatHandler(chatServiceStub()),
		handler.NewTaskHandler(taskServiceStub()),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/not-a-number", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got %d want %d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestListTasksRoute(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	engine := New(
		handler.NewChatHandler(chatServiceStub()),
		handler.NewTaskHandler(taskServiceStub()),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?user_id=42", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got %d want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestListTasksRouteRejectsInvalidUserID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	engine := New(
		handler.NewChatHandler(chatServiceStub()),
		handler.NewTaskHandler(taskServiceStub()),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?user_id=abc", nil)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got %d want %d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

type chatServiceContract interface {
	Handle(ctx context.Context, req chatapp.Request) (chatapp.Response, error)
}

type taskServiceContract interface {
	CreateAsyncTask(ctx context.Context, task taskdomain.Task) error
	GetTask(ctx context.Context, id int64) (taskdomain.Task, error)
	ListTasks(ctx context.Context, userID int64, limit int) ([]taskdomain.Task, error)
}

func chatServiceStub() chatServiceContract {
	return chatapp.NewService(planner.NewKeywordPlanner(), executor.NewStubExecutor(), nil)
}

func taskServiceStub() taskServiceContract {
	return taskapp.NewService(&routerTaskRepo{
		task: taskdomain.Task{
			ID:       1001,
			UserID:   42,
			TaskName: "query_course",
			Status:   taskdomain.StatusSuccess,
			Result:   "done",
		},
	}, producer.NewInMemoryProducer())
}

type routerTaskRepo struct {
	task  taskdomain.Task
	tasks []taskdomain.Task
}

func (r *routerTaskRepo) Save(_ context.Context, task taskdomain.Task) error {
	r.task = task
	return nil
}

func (r *routerTaskRepo) UpdateStatus(_ context.Context, _ int64, status taskdomain.Status, result string) error {
	r.task.Status = status
	r.task.Result = result
	return nil
}

func (r *routerTaskRepo) FindByID(_ context.Context, id int64) (taskdomain.Task, error) {
	task := r.task
	task.ID = id
	return task, nil
}

func (r *routerTaskRepo) ListByUser(_ context.Context, userID int64, limit int) ([]taskdomain.Task, error) {
	return r.tasks, nil
}
