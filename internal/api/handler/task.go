package handler

import (
	"context"
	"net/http"
	"strconv"

	taskdomain "campus-agent/internal/domain/task"
	"campus-agent/pkg/response"
	"github.com/gin-gonic/gin"
)

type TaskService interface {
	CreateAsyncTask(ctx context.Context, task taskdomain.Task) error
	GetTask(ctx context.Context, id int64) (taskdomain.Task, error)
	ListTasks(ctx context.Context, userID int64, limit int) ([]taskdomain.Task, error)
}

type TaskHandler struct {
	service TaskService
}

type createTaskRequest struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	TaskName string `json:"task_name"`
}

func NewTaskHandler(service TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid request body"))
		return
	}

	task := taskdomain.Task{
		ID:       req.ID,
		UserID:   req.UserID,
		TaskName: req.TaskName,
		Status:   taskdomain.StatusPending,
	}

	if err := h.service.CreateAsyncTask(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusAccepted, response.Success(task))
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid task id"))
		return
	}

	task, err := h.service.GetTask(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(task))
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid user id"))
		return
	}

	tasks, err := h.service.ListTasks(c.Request.Context(), userID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(tasks))
}
