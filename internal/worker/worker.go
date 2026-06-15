package worker

import (
	"campus-agent/internal/agent/executor"
	taskapp "campus-agent/internal/app/task"
	reminderdomain "campus-agent/internal/domain/reminder"
	taskdomain "campus-agent/internal/domain/task"
	"campus-agent/internal/mq/consumer"
	"campus-agent/internal/mq/message"
	coursetool "campus-agent/internal/tool/course"
	knowledgetool "campus-agent/internal/tool/knowledge"
	remindertool "campus-agent/internal/tool/reminder"
)

func NewTaskExecutionConsumer(source consumer.AMQPDeliverySource, repo taskdomain.Repository, executorAgent executor.Executor) *consumer.RabbitMQConsumer {
	handler := taskapp.NewExecutionHandler(repo, executorAgent)
	return consumer.NewRabbitMQConsumer(source, message.QueueTaskExecute, handler)
}

func NewDefaultToolExecutor(reminderRepo reminderdomain.Repository, knowledgeDocs []knowledgetool.Document) executor.Executor {
	return executor.NewToolExecutor(
		coursetool.NewStubTool(),
		remindertool.NewRepositoryTool(reminderRepo),
		knowledgetool.NewLocalTool(knowledgeDocs),
	)
}
