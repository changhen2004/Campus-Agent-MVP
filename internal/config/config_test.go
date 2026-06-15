package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParsesYAMLConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := []byte(`
server:
  host: "127.0.0.1"
  port: 9090
mysql:
  host: "mysql"
  port: 3306
  user: "campus"
  password: "campus"
  database: "campus_agent"
redis:
  addr: "redis:6379"
  password: ""
  db: 1
rabbitmq:
  addr: "amqp://guest:guest@rabbitmq:5672/"
  exchange: "campus.agent"
  task_execute_queue: "task.execute"
  task_notification_queue: "task.notification"
llm:
  provider: "openai-compatible"
  endpoint: "https://api.example.com/v1"
  api_key: "secret"
  model: "deepseek-chat"
rag:
  provider: "qdrant"
  endpoint: "http://qdrant:6333"
  collection: "campus_knowledge"
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Fatalf("server host mismatch: got %q", cfg.Server.Host)
	}

	if cfg.Server.Port != 9090 {
		t.Fatalf("server port mismatch: got %d", cfg.Server.Port)
	}

	if cfg.RabbitMQ.Exchange != "campus.agent" {
		t.Fatalf("exchange mismatch: got %q", cfg.RabbitMQ.Exchange)
	}
}
