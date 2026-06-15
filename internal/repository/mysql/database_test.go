package mysql

import (
	"strings"
	"testing"

	"campus-agent/internal/config"
)

func TestBuildDSN(t *testing.T) {
	t.Parallel()

	dsn := BuildDSN(config.MySQLConfig{
		Host:     "mysql",
		Port:     3306,
		User:     "campus",
		Password: "campus",
		Database: "campus_agent",
	})

	wants := []string{
		"campus:campus@tcp(mysql:3306)/campus_agent",
		"charset=utf8mb4",
		"parseTime=True",
		"loc=Local",
	}

	for _, want := range wants {
		if !strings.Contains(dsn, want) {
			t.Fatalf("dsn %q does not contain %q", dsn, want)
		}
	}
}
