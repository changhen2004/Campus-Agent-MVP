package producer

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"campus-agent/internal/mq/message"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPPublisher interface {
	PublishWithContext(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
}

type InMemoryProducer struct {
	mu       sync.Mutex
	messages []message.TaskMessage
}

func NewInMemoryProducer() *InMemoryProducer {
	return &InMemoryProducer{}
}

func (p *InMemoryProducer) Publish(_ context.Context, msg message.TaskMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.messages = append(p.messages, msg)
	return nil
}

func (p *InMemoryProducer) Snapshot() []message.TaskMessage {
	p.mu.Lock()
	defer p.mu.Unlock()

	out := make([]message.TaskMessage, len(p.messages))
	copy(out, p.messages)
	return out
}

type RabbitMQProducer struct {
	publisher AMQPPublisher
	exchange  string
}

func NewRabbitMQProducer(publisher AMQPPublisher, exchange string) *RabbitMQProducer {
	return &RabbitMQProducer{
		publisher: publisher,
		exchange:  exchange,
	}
}

func (p *RabbitMQProducer) Publish(ctx context.Context, msg message.TaskMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.publisher.PublishWithContext(ctx, p.exchange, message.QueueTaskExecute, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	})
}
