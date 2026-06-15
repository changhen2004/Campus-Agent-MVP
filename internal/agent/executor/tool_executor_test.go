package executor

import (
	"context"
	"testing"

	"campus-agent/internal/agent/planner"
	coursedomain "campus-agent/internal/domain/course"
	reminderdomain "campus-agent/internal/domain/reminder"
)

func TestToolExecutorQueriesCourse(t *testing.T) {
	t.Parallel()

	courseTool := &fakeCourseTool{
		courses: []coursedomain.Course{
			{Name: "Database Systems", Location: "A201"},
		},
	}
	executorAgent := NewToolExecutor(courseTool, nil, nil)

	results, err := executorAgent.Execute(context.Background(), []planner.Task{
		{
			Name:   planner.TaskQueryCourse,
			UserID: 42,
			Input:  "帮我查询明天课程",
		},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if courseTool.userID != 42 {
		t.Fatalf("course user mismatch: got %d want %d", courseTool.userID, 42)
	}

	if len(results) != 1 || results[0].Status != StatusSuccess {
		t.Fatalf("unexpected results: %+v", results)
	}

	if results[0].Output == "" {
		t.Fatal("expected course output")
	}
}

func TestToolExecutorCreatesReminder(t *testing.T) {
	t.Parallel()

	reminderTool := &fakeReminderTool{}
	executorAgent := NewToolExecutor(nil, reminderTool, nil)

	results, err := executorAgent.Execute(context.Background(), []planner.Task{
		{
			Name:   planner.TaskCreateReminder,
			UserID: 7,
			Input:  "提醒我完成数据库实验",
		},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if reminderTool.reminder.UserID != 7 {
		t.Fatalf("reminder user mismatch: got %d want %d", reminderTool.reminder.UserID, 7)
	}

	if reminderTool.reminder.Title == "" || reminderTool.reminder.Content == "" {
		t.Fatalf("expected reminder title and content: %+v", reminderTool.reminder)
	}

	if len(results) != 1 || results[0].Status != StatusSuccess {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestToolExecutorSearchesKnowledge(t *testing.T) {
	t.Parallel()

	knowledgeTool := &fakeKnowledgeTool{
		results: []string{"实验报告通过教务平台提交"},
	}
	executorAgent := NewToolExecutor(nil, nil, knowledgeTool)

	results, err := executorAgent.Execute(context.Background(), []planner.Task{
		{
			Name:  planner.TaskSearchKnowledge,
			Input: "实验报告怎么提交？",
		},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if knowledgeTool.query != "实验报告怎么提交？" {
		t.Fatalf("knowledge query mismatch: got %q", knowledgeTool.query)
	}

	if len(results) != 1 || results[0].Output == "" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

type fakeCourseTool struct {
	userID  int64
	courses []coursedomain.Course
}

func (t *fakeCourseTool) QueryTomorrowCourses(_ context.Context, userID int64) ([]coursedomain.Course, error) {
	t.userID = userID
	return t.courses, nil
}

type fakeReminderTool struct {
	reminder reminderdomain.Reminder
}

func (t *fakeReminderTool) Create(_ context.Context, reminder reminderdomain.Reminder) error {
	t.reminder = reminder
	return nil
}

type fakeKnowledgeTool struct {
	query   string
	results []string
}

func (t *fakeKnowledgeTool) Search(_ context.Context, query string) ([]string, error) {
	t.query = query
	return t.results, nil
}
