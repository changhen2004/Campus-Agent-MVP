package server

import (
	"campus-agent/internal/agent/executor"
	"campus-agent/internal/agent/planner"
	chatapp "campus-agent/internal/app/chat"
	taskapp "campus-agent/internal/app/task"
	"campus-agent/internal/config"
	reminderdomain "campus-agent/internal/domain/reminder"
	taskdomain "campus-agent/internal/domain/task"
	"campus-agent/internal/mq/producer"
	coursetool "campus-agent/internal/tool/course"
	knowledgetool "campus-agent/internal/tool/knowledge"
	remindertool "campus-agent/internal/tool/reminder"
)

type Services struct {
	Chat *chatapp.Service
	Task *taskapp.Service
}

func NewServices(taskRepo taskdomain.Repository, reminderRepo reminderdomain.Repository, publisher producer.AMQPPublisher, knowledgeDocs []knowledgetool.Document, cfg config.RabbitMQConfig) Services {
	plannerAgent := planner.NewKeywordPlanner()
	executorAgent := executor.NewToolExecutor(
		coursetool.NewStubTool(),
		remindertool.NewRepositoryTool(reminderRepo),
		knowledgetool.NewLocalTool(knowledgeDocs),
	)
	taskProducer := producer.NewRabbitMQProducer(publisher, cfg.Exchange)

	return Services{
		Chat: chatapp.NewService(plannerAgent, executorAgent, nil),
		Task: taskapp.NewService(taskRepo, taskProducer),
	}
}
