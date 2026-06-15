package consumer

import (
	"context"
	"encoding/json"
	"testing"

	"campus-agent/internal/mq/message"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestRabbitMQConsumerHandlesValidDeliveryAndAcks(t *testing.T) {
	t.Parallel()

	taskMsg := message.TaskMessage{
		TaskID:   1001,
		UserID:   42,
		TaskName: "query_course",
	}
	body, err := json.Marshal(taskMsg)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	acker := &fakeAcknowledger{}
	deliveries := make(chan amqp.Delivery, 1)
	deliveries <- amqp.Delivery{
		Body:         body,
		DeliveryTag:  1,
		Acknowledger: acker,
	}
	close(deliveries)

	handler := &fakeHandler{}
	consumer := NewRabbitMQConsumer(&fakeDeliverySource{deliveries: deliveries}, message.QueueTaskExecute, handler)

	if err := consumer.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	if len(handler.messages) != 1 || handler.messages[0] != taskMsg {
		t.Fatalf("handler messages mismatch: got %+v want %+v", handler.messages, taskMsg)
	}

	if acker.acks != 1 {
		t.Fatalf("ack count mismatch: got %d want %d", acker.acks, 1)
	}
}

func TestRabbitMQConsumerNacksInvalidJSON(t *testing.T) {
	t.Parallel()

	acker := &fakeAcknowledger{}
	deliveries := make(chan amqp.Delivery, 1)
	deliveries <- amqp.Delivery{
		Body:         []byte("{"),
		DeliveryTag:  1,
		Acknowledger: acker,
	}
	close(deliveries)

	handler := &fakeHandler{}
	consumer := NewRabbitMQConsumer(&fakeDeliverySource{deliveries: deliveries}, message.QueueTaskExecute, handler)

	if err := consumer.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	if len(handler.messages) != 0 {
		t.Fatalf("handler should not be called for invalid JSON: got %+v", handler.messages)
	}

	if acker.nacks != 1 {
		t.Fatalf("nack count mismatch: got %d want %d", acker.nacks, 1)
	}

	if acker.requeue {
		t.Fatal("invalid JSON should not be requeued")
	}
}

type fakeDeliverySource struct {
	deliveries <-chan amqp.Delivery
}

func (s *fakeDeliverySource) Consume(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return s.deliveries, nil
}

type fakeHandler struct {
	messages []message.TaskMessage
	err      error
}

func (h *fakeHandler) Handle(_ context.Context, msg message.TaskMessage) error {
	h.messages = append(h.messages, msg)
	return h.err
}

type fakeAcknowledger struct {
	acks    int
	nacks   int
	requeue bool
}

func (a *fakeAcknowledger) Ack(_ uint64, _ bool) error {
	a.acks++
	return nil
}

func (a *fakeAcknowledger) Nack(_ uint64, _ bool, requeue bool) error {
	a.nacks++
	a.requeue = requeue
	return nil
}

func (a *fakeAcknowledger) Reject(_ uint64, requeue bool) error {
	a.nacks++
	a.requeue = requeue
	return nil
}
