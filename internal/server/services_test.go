package server

import (
	"context"
	"encoding/json"
	"testing"

	chatapp "campus-agent/internal/app/chat"
	"campus-agent/internal/config"
	reminderdomain "campus-agent/internal/domain/reminder"
	taskdomain "campus-agent/internal/domain/task"
	"campus-agent/internal/mq/message"
	knowledgetool "campus-agent/internal/tool/knowledge"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestNewServicesWiresTaskServiceToRabbitMQ(t *testing.T) {
	t.Parallel()

	repo := &fakeTaskRepo{}
	reminderRepo := &fakeReminderRepo{}
	publisher := &fakePublisher{}
	services := NewServices(repo, reminderRepo, publisher, nil, config.RabbitMQConfig{
		Exchange: message.ExchangeCampusAgent,
	})

	task := taskdomain.Task{
		ID:       1001,
		UserID:   42,
		TaskName: string(message.QueueTaskExecute),
		Status:   taskdomain.StatusPending,
	}

	if err := services.Task.CreateAsyncTask(context.Background(), task); err != nil {
		t.Fatalf("CreateAsyncTask returned error: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("saved task count mismatch: got %d want %d", len(repo.saved), 1)
	}

	if publisher.exchange != message.ExchangeCampusAgent {
		t.Fatalf("exchange mismatch: got %q want %q", publisher.exchange, message.ExchangeCampusAgent)
	}

	if publisher.key != message.QueueTaskExecute {
		t.Fatalf("routing key mismatch: got %q want %q", publisher.key, message.QueueTaskExecute)
	}

	var decoded message.TaskMessage
	if err := json.Unmarshal(publisher.body, &decoded); err != nil {
		t.Fatalf("published body is not JSON: %v", err)
	}

	if decoded.TaskID != task.ID || decoded.UserID != task.UserID || decoded.TaskName != task.TaskName {
		t.Fatalf("published message mismatch: got %+v", decoded)
	}
}

func TestNewServicesWiresChatReminderToRepository(t *testing.T) {
	t.Parallel()

	reminderRepo := &fakeReminderRepo{}
	services := NewServices(&fakeTaskRepo{}, reminderRepo, &fakePublisher{}, nil, config.RabbitMQConfig{
		Exchange: message.ExchangeCampusAgent,
	})

	_, err := services.Chat.Handle(context.Background(), chatRequest(42, "提醒我完成数据库实验"))
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if len(reminderRepo.saved) != 1 {
		t.Fatalf("saved reminder count mismatch: got %d want %d", len(reminderRepo.saved), 1)
	}

	if reminderRepo.saved[0].UserID != 42 {
		t.Fatalf("reminder user mismatch: got %d want %d", reminderRepo.saved[0].UserID, 42)
	}
}

func TestNewServicesWiresKnowledgeDocuments(t *testing.T) {
	t.Parallel()

	services := NewServices(&fakeTaskRepo{}, &fakeReminderRepo{}, &fakePublisher{}, []knowledgetool.Document{
		{
			ID:      "lab-report",
			Title:   "实验报告提交",
			Content: "实验报告需要通过教务平台提交。",
		},
	}, config.RabbitMQConfig{
		Exchange: message.ExchangeCampusAgent,
	})

	resp, err := services.Chat.Handle(context.Background(), chatRequest(42, "实验报告怎么提交"))
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("result count mismatch: got %d want %d", len(resp.Results), 1)
	}

	if resp.Results[0].Output == "" {
		t.Fatal("expected knowledge output")
	}
}

type fakeTaskRepo struct {
	saved []taskdomain.Task
}

func (r *fakeTaskRepo) Save(_ context.Context, task taskdomain.Task) error {
	r.saved = append(r.saved, task)
	return nil
}

func (r *fakeTaskRepo) UpdateStatus(_ context.Context, _ int64, _ taskdomain.Status, _ string) error {
	return nil
}

func (r *fakeTaskRepo) FindByID(_ context.Context, id int64) (taskdomain.Task, error) {
	return taskdomain.Task{ID: id}, nil
}

func (r *fakeTaskRepo) ListByUser(_ context.Context, userID int64, limit int) ([]taskdomain.Task, error) {
	return r.saved, nil
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

func chatRequest(userID int64, message string) chatapp.Request {
	return chatapp.Request{
		UserID:  userID,
		Message: message,
	}
}

type fakePublisher struct {
	exchange string
	key      string
	body     []byte
}

func (p *fakePublisher) PublishWithContext(_ context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
	p.exchange = exchange
	p.key = key
	p.body = msg.Body
	return nil
}
