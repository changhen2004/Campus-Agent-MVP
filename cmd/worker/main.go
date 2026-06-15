package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"campus-agent/internal/config"
	localknowledge "campus-agent/internal/knowledge/local"
	"campus-agent/internal/mq/topology"
	"campus-agent/internal/repository/mysql"
	"campus-agent/internal/runtime/retry"
	"campus-agent/internal/worker"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	executorAgent := worker.NewDefaultToolExecutor(reminderRepo, knowledgeDocs)
	taskConsumer := worker.NewTaskExecutionConsumer(channel, taskRepo, executorAgent)

	log.Println("campus-agent worker consuming task.execute")
	if err := taskConsumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("worker stopped with error: %v", err)
	}
}
