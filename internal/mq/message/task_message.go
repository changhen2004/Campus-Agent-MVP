package message

const (
	ExchangeCampusAgent   = "campus.agent"
	QueueTaskExecute      = "task.execute"
	QueueTaskNotification = "task.notification"
)

type TaskMessage struct {
	TaskID   int64  `json:"task_id"`
	UserID   int64  `json:"user_id"`
	TaskName string `json:"task_name"`
}
