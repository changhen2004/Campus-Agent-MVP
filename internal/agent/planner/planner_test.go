package planner

import (
	"context"
	"testing"
)

func TestPlan(t *testing.T) {
	t.Parallel()

	plannerAgent := NewKeywordPlanner()

	tests := []struct {
		name       string
		message    string
		wantIntent Intent
		wantTasks  int
	}{
		{
			name:       "course and reminder",
			message:    "帮我查询明天课程并提醒我完成数据库实验",
			wantIntent: IntentQueryCourseCreateReminder,
			wantTasks:  2,
		},
		{
			name:       "knowledge query",
			message:    "实验报告怎么提交？",
			wantIntent: IntentKnowledgeQuery,
			wantTasks:  1,
		},
		{
			name:       "empty input",
			message:    "",
			wantIntent: IntentGeneral,
			wantTasks:  0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := plannerAgent.Plan(context.Background(), tt.message)
			if err != nil {
				t.Fatalf("Plan returned error: %v", err)
			}

			if result.Intent != tt.wantIntent {
				t.Fatalf("intent mismatch: got %q want %q", result.Intent, tt.wantIntent)
			}

			if len(result.Tasks) != tt.wantTasks {
				t.Fatalf("task count mismatch: got %d want %d", len(result.Tasks), tt.wantTasks)
			}
		})
	}
}
