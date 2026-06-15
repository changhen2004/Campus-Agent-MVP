package planner

import (
	"context"
	"strings"
)

type Intent string

const (
	IntentGeneral                   Intent = "general_question"
	IntentQueryCourse               Intent = "query_course"
	IntentCreateReminder            Intent = "create_reminder"
	IntentKnowledgeQuery            Intent = "knowledge_query"
	IntentQueryCourseCreateReminder Intent = "query_course_and_create_reminder"
)

type TaskName string

const (
	TaskQueryCourse     TaskName = "query_course"
	TaskCreateReminder  TaskName = "create_reminder"
	TaskSearchKnowledge TaskName = "search_knowledge"
)

type Task struct {
	Name   TaskName
	UserID int64
	Input  string
}

type Result struct {
	Intent Intent
	Tasks  []Task
}

type Planner interface {
	Plan(ctx context.Context, message string) (Result, error)
}

type KeywordPlanner struct{}

func NewKeywordPlanner() *KeywordPlanner {
	return &KeywordPlanner{}
}

func (p *KeywordPlanner) Plan(_ context.Context, message string) (Result, error) {
	text := strings.TrimSpace(message)
	if text == "" {
		return Result{Intent: IntentGeneral}, nil
	}

	hasCourse := strings.Contains(text, "课程")
	hasReminder := strings.Contains(text, "提醒")
	hasKnowledge := strings.Contains(text, "实验报告") ||
		strings.Contains(text, "规章") ||
		strings.Contains(text, "提交")

	switch {
	case hasCourse && hasReminder:
		return Result{
			Intent: IntentQueryCourseCreateReminder,
			Tasks: []Task{
				{Name: TaskQueryCourse},
				{Name: TaskCreateReminder},
			},
		}, nil
	case hasCourse:
		return Result{
			Intent: IntentQueryCourse,
			Tasks:  []Task{{Name: TaskQueryCourse}},
		}, nil
	case hasKnowledge:
		return Result{
			Intent: IntentKnowledgeQuery,
			Tasks:  []Task{{Name: TaskSearchKnowledge}},
		}, nil
	case hasReminder:
		return Result{
			Intent: IntentCreateReminder,
			Tasks:  []Task{{Name: TaskCreateReminder}},
		}, nil
	default:
		return Result{Intent: IntentGeneral}, nil
	}
}
