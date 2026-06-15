package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"campus-agent/internal/api/handler"
	"campus-agent/internal/api/router"
	"campus-agent/internal/config"
	localknowledge "campus-agent/internal/knowledge/local"
	"campus-agent/internal/mq/topology"
	"campus-agent/internal/repository/mysql"
	"campus-agent/internal/runtime/retry"
	"campus-agent/internal/server"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

func main() {
	configPath := os.Getenv("CAMPUS_AGENT_CONFIG")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	retryOptions := retry.Options{
		Attempts: 30,
		Delay:    2 * time.Second,
	}

	var db *gorm.DB
	if err := retry.Do(ctx, retryOptions, func(context.Context) error {
		var err error
		db, err = mysql.Open(cfg.MySQL)
		return err
	}); err != nil {
		log.Fatalf("open mysql after retries: %v", err)
	}

	if err := mysql.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate mysql: %v", err)
	}

	var conn *amqp.Connection
	if err := retry.Do(ctx, retryOptions, func(context.Context) error {
		var err error
		conn, err = amqp.Dial(cfg.RabbitMQ.Addr)
		return err
	}); err != nil {
		log.Fatalf("dial rabbitmq after retries: %v", err)
	}
	defer conn.Close()

	var channel *amqp.Channel
	if err := retry.Do(ctx, retryOptions, func(context.Context) error {
		var err error
		channel, err = conn.Channel()
		return err
	}); err != nil {
		log.Fatalf("open rabbitmq channel after retries: %v", err)
	}
	defer channel.Close()

	if err := topology.DeclareTaskExecution(channel, cfg.RabbitMQ); err != nil {
		log.Fatalf("declare rabbitmq topology: %v", err)
	}

	taskRepo := mysql.NewTaskRepository(db)
	reminderRepo := mysql.NewReminderRepository(db)
	knowledgeDocs, err := localknowledge.LoadMarkdownDir("docs/knowledge")
	if err != nil {
		log.Fatalf("load knowledge docs: %v", err)
	}
	services := server.NewServices(taskRepo, reminderRepo, channel, knowledgeDocs, cfg.RabbitMQ)

	chatHandler := handler.NewChatHandler(services.Chat)
	taskHandler := handler.NewTaskHandler(services.Task)
	staticFS := http.FS(os.DirFS("web/static"))

	mux := router.New(chatHandler, taskHandler, staticFS)

	log.Printf("campus-agent server listening on %s", cfg.HTTPAddr())
	if err := mux.Run(cfg.HTTPAddr()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
