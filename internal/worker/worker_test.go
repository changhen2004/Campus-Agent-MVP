package worker

import (
	"context"
	"encoding/json"
	"testing"

	"campus-agent/internal/agent/executor"
	"campus-agent/internal/agent/planner"
	reminderdomain "campus-agent/internal/domain/reminder"
	taskdomain "campus-agent/internal/domain/task"
	"campus-agent/internal/mq/message"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestNewTaskExecutionConsumerWiresTaskExecuteQueue(t *testing.T) {
	t.Parallel()

	taskMsg := message.TaskMessage{
		TaskID:   1001,
		UserID:   42,
		TaskName: string(planner.TaskQueryCourse),
	}
	body, err := json.Marshal(taskMsg)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	source := &fakeSource{
		deliveries: make(chan amqp.Delivery, 1),
	}
	source.deliveries <- amqp.Delivery{
		Body:         body,
		DeliveryTag:  1,
		Acknowledger: &fakeAcker{},
	}
	close(source.deliveries)

	repo := &fakeRepo{}
	consumer := NewTaskExecutionConsumer(source, repo, executor.NewStubExecutor())

	if err := consumer.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	if source.queue != message.QueueTaskExecute {
		t.Fatalf("queue mismatch: got %q want %q", source.queue, message.QueueTaskExecute)
	}

	if len(repo.updates) != 2 {
		t.Fatalf("update count mismatch: got %d want %d", len(repo.updates), 2)
	}

	if repo.updates[0].status != taskdomain.StatusRunning {
		t.Fatalf("first status mismatch: got %q want %q", repo.updates[0].status, taskdomain.StatusRunning)
	}

	if repo.updates[1].status != taskdomain.StatusSuccess {
		t.Fatalf("second status mismatch: got %q want %q", repo.updates[1].status, taskdomain.StatusSuccess)
	}
}

func TestNewDefaultToolExecutorPersistsReminder(t *testing.T) {
	t.Parallel()

	reminderRepo := &fakeReminderRepo{}
	executorAgent := NewDefaultToolExecutor(reminderRepo, nil)

	_, err := executorAgent.Execute(context.Background(), []planner.Task{
		{
			Name:   planner.TaskCreateReminder,
			UserID: 42,
			Input:  "提醒我完成数据库实验",
		},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(reminderRepo.saved) != 1 {
		t.Fatalf("saved reminder count mismatch: got %d want %d", len(reminderRepo.saved), 1)
	}
}

type fakeSource struct {
	queue      string
	deliveries chan amqp.Delivery
}

func (s *fakeSource) Consume(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	s.queue = queue
	return s.deliveries, nil
}

type fakeRepo struct {
	updates []taskUpdate
}

type taskUpdate struct {
	status taskdomain.Status
	result string
}

func (r *fakeRepo) Save(_ context.Context, _ taskdomain.Task) error {
	return nil
}

func (r *fakeRepo) UpdateStatus(_ context.Context, _ int64, status taskdomain.Status, result string) error {
	r.updates = append(r.updates, taskUpdate{
		status: status,
		result: result,
	})
	return nil
}

func (r *fakeRepo) FindByID(_ context.Context, id int64) (taskdomain.Task, error) {
	return taskdomain.Task{ID: id}, nil
}

func (r *fakeRepo) ListByUser(_ context.Context, userID int64, limit int) ([]taskdomain.Task, error) {
	return []taskdomain.Task{}, nil
}

type fakeAcker struct{}

func (a *fakeAcker) Ack(_ uint64, _ bool) error {
	return nil
}

func (a *fakeAcker) Nack(_ uint64, _ bool, _ bool) error {
	return nil
}

func (a *fakeAcker) Reject(_ uint64, _ bool) error {
	return nil
}

type fakeReminderRepo struct {
	saved []reminderdomain.Reminder
}

func (r *fakeReminderRepo) Save(_ context.Context, reminder reminderdomain.Reminder) error {
	r.saved = append(r.saved, reminder)
	return nil
}

func (r *fakeReminderRepo) FindByID(_ context.Context, id int64) (reminderdomain.Reminder, error) {
	return reminderdomain.Reminder{ID: id}, nil
}
