package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"campus-agent/internal/agent/planner"
	reminderdomain "campus-agent/internal/domain/reminder"
	coursetool "campus-agent/internal/tool/course"
	knowledgetool "campus-agent/internal/tool/knowledge"
	remindertool "campus-agent/internal/tool/reminder"
)

type Status string

const (
	StatusSuccess Status = "success"
)

type TaskResult struct {
	Task   planner.TaskName
	Status Status
	Output string
}

type Executor interface {
	Execute(ctx context.Context, tasks []planner.Task) ([]TaskResult, error)
}

type StubExecutor struct{}

func NewStubExecutor() *StubExecutor {
	return &StubExecutor{}
}

type ToolExecutor struct {
	courseTool    coursetool.Tool
	reminderTool  remindertool.Tool
	knowledgeTool knowledgetool.Tool
}

func NewToolExecutor(courseTool coursetool.Tool, reminderTool remindertool.Tool, knowledgeTool knowledgetool.Tool) *ToolExecutor {
	return &ToolExecutor{
		courseTool:    courseTool,
		reminderTool:  reminderTool,
		knowledgeTool: knowledgeTool,
	}
}

func (e *ToolExecutor) Execute(ctx context.Context, tasks []planner.Task) ([]TaskResult, error) {
	results := make([]TaskResult, 0, len(tasks))
	for _, task := range tasks {
		result, err := e.executeOne(ctx, task)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (e *ToolExecutor) executeOne(ctx context.Context, task planner.Task) (TaskResult, error) {
	switch task.Name {
	case planner.TaskQueryCourse:
		return e.executeCourse(ctx, task)
	case planner.TaskCreateReminder:
		return e.executeReminder(ctx, task)
	case planner.TaskSearchKnowledge:
		return e.executeKnowledge(ctx, task)
	default:
		return TaskResult{}, fmt.Errorf("unsupported task: %s", task.Name)
	}
}

func (e *ToolExecutor) executeCourse(ctx context.Context, task planner.Task) (TaskResult, error) {
	if e.courseTool == nil {
		return TaskResult{}, fmt.Errorf("course tool is not configured")
	}

	courses, err := e.courseTool.QueryTomorrowCourses(ctx, task.UserID)
	if err != nil {
		return TaskResult{}, err
	}

	output, err := json.Marshal(courses)
	if err != nil {
		return TaskResult{}, err
	}

	return TaskResult{
		Task:   task.Name,
		Status: StatusSuccess,
		Output: string(output),
	}, nil
}

func (e *ToolExecutor) executeReminder(ctx context.Context, task planner.Task) (TaskResult, error) {
	if e.reminderTool == nil {
		return TaskResult{}, fmt.Errorf("reminder tool is not configured")
	}

	reminder := reminderdomain.Reminder{
		UserID:  task.UserID,
		Title:   reminderTitle(task.Input),
		Content: task.Input,
	}
	if err := e.reminderTool.Create(ctx, reminder); err != nil {
		return TaskResult{}, err
	}

	return TaskResult{
		Task:   task.Name,
		Status: StatusSuccess,
		Output: "reminder created",
	}, nil
}

func (e *ToolExecutor) executeKnowledge(ctx context.Context, task planner.Task) (TaskResult, error) {
	if e.knowledgeTool == nil {
		return TaskResult{}, fmt.Errorf("knowledge tool is not configured")
	}

	documents, err := e.knowledgeTool.Search(ctx, task.Input)
	if err != nil {
		return TaskResult{}, err
	}

	return TaskResult{
		Task:   task.Name,
		Status: StatusSuccess,
		Output: strings.Join(documents, "\n"),
	}, nil
}

func reminderTitle(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return "校园提醒"
	}

	runes := []rune(input)
	if len(runes) > 20 {
		return string(runes[:20])
	}
	return input
}

func (e *StubExecutor) Execute(_ context.Context, tasks []planner.Task) ([]TaskResult, error) {
	results := make([]TaskResult, 0, len(tasks))
	for _, task := range tasks {
		results = append(results, TaskResult{
			Task:   task.Name,
			Status: StatusSuccess,
			Output: fmt.Sprintf("%s executed by stub executor", task.Name),
		})
	}
	return results, nil
}
