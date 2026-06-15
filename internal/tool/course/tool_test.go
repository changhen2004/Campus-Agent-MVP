package course

import (
	"context"
	"testing"

	coursedomain "campus-agent/internal/domain/course"
)

func TestStaticToolReturnsConfiguredCourses(t *testing.T) {
	t.Parallel()

	tool := NewStaticTool([]coursedomain.Course{
		{
			Name:     "Database Systems",
			Location: "A201",
		},
	})

	courses, err := tool.QueryTomorrowCourses(context.Background(), 42)
	if err != nil {
		t.Fatalf("QueryTomorrowCourses returned error: %v", err)
	}

	if len(courses) != 1 {
		t.Fatalf("course count mismatch: got %d want %d", len(courses), 1)
	}

	if courses[0].Name != "Database Systems" {
		t.Fatalf("course name mismatch: got %q", courses[0].Name)
	}
}
