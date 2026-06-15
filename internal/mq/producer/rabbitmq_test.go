package producer

import (
	"context"
	"encoding/json"
	"testing"

	"campus-agent/internal/mq/message"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestRabbitMQProducerPublishesTaskMessage(t *testing.T) {
	t.Parallel()

	channel := &fakePublisher{}
	producer := NewRabbitMQProducer(channel, message.ExchangeCampusAgent)

	msg := message.TaskMessage{
		TaskID:   1001,
		UserID:   42,
		TaskName: "query_course",
	}

	if err := producer.Publish(context.Background(), msg); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	if channel.exchange != message.ExchangeCampusAgent {
		t.Fatalf("exchange mismatch: got %q want %q", channel.exchange, message.ExchangeCampusAgent)
	}

	if channel.key != message.QueueTaskExecute {
		t.Fatalf("routing key mismatch: got %q want %q", channel.key, message.QueueTaskExecute)
	}

	var decoded message.TaskMessage
	if err := json.Unmarshal(channel.publishing.Body, &decoded); err != nil {
		t.Fatalf("published body is not valid JSON: %v", err)
	}

	if decoded != msg {
		t.Fatalf("published message mismatch: got %+v want %+v", decoded, msg)
	}

	if channel.publishing.ContentType != "application/json" {
		t.Fatalf("content type mismatch: got %q", channel.publishing.ContentType)
	}
}

type fakePublisher struct {
	exchange   string
	key        string
	mandatory  bool
	immediate  bool
	publishing amqp.Publishing
}

func (f *fakePublisher) PublishWithContext(_ context.Context, exchange string, key string, mandatory bool, immediate bool, publishing amqp.Publishing) error {
	f.exchange = exchange
	f.key = key
	f.mandatory = mandatory
	f.immediate = immediate
	f.publishing = publishing
	return nil
}
