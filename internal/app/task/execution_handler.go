package taskapp

import (
	"context"
	"encoding/json"

	"campus-agent/internal/agent/executor"
	"campus-agent/internal/agent/planner"
	taskdomain "campus-agent/internal/domain/task"
	"campus-agent/internal/mq/message"
)

type ExecutionHandler struct {
	repo     taskdomain.Repository
	executor executor.Executor
}

func NewExecutionHandler(repo taskdomain.Repository, executorAgent executor.Executor) *ExecutionHandler {
	return &ExecutionHandler{
		repo:     repo,
		executor: executorAgent,
	}
}

func (h *ExecutionHandler) Handle(ctx context.Context, msg message.TaskMessage) error {
	if err := h.repo.UpdateStatus(ctx, msg.TaskID, taskdomain.StatusRunning, ""); err != nil {
		return err
	}

	results, err := h.executor.Execute(ctx, []planner.Task{
		{
			Name:   planner.TaskName(msg.TaskName),
			UserID: msg.UserID,
			Input:  msg.TaskName,
		},
	})
	if err != nil {
		_ = h.repo.UpdateStatus(ctx, msg.TaskID, taskdomain.StatusFailed, err.Error())
		return err
	}

	resultBody, err := json.Marshal(results)
	if err != nil {
		_ = h.repo.UpdateStatus(ctx, msg.TaskID, taskdomain.StatusFailed, err.Error())
		return err
	}

	return h.repo.UpdateStatus(ctx, msg.TaskID, taskdomain.StatusSuccess, string(resultBody))
}
