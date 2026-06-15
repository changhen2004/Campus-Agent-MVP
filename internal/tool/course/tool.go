package course

import (
	"context"

	coursedomain "campus-agent/internal/domain/course"
)

type Tool interface {
	QueryTomorrowCourses(ctx context.Context, userID int64) ([]coursedomain.Course, error)
}

type StaticTool struct {
	courses []coursedomain.Course
}

func NewStaticTool(courses []coursedomain.Course) *StaticTool {
	copied := make([]coursedomain.Course, len(courses))
	copy(copied, courses)
	return &StaticTool{courses: copied}
}

func (t *StaticTool) QueryTomorrowCourses(_ context.Context, _ int64) ([]coursedomain.Course, error) {
	copied := make([]coursedomain.Course, len(t.courses))
	copy(copied, t.courses)
	return copied, nil
}

type StubTool struct{}

func NewStubTool() *StubTool {
	return &StubTool{}
}

func (t *StubTool) QueryTomorrowCourses(_ context.Context, _ int64) ([]coursedomain.Course, error) {
	return NewStaticTool([]coursedomain.Course{
		{
			Name:      "Database Systems",
			Location:  "Teaching Building A201",
			Teacher:   "TBD",
			Weekday:   "Tomorrow",
			StartTime: "08:00",
			EndTime:   "09:40",
		},
	}).QueryTomorrowCourses(context.Background(), 0)
}
