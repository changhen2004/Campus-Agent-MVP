package consumer

import (
	"context"
	"encoding/json"

	"campus-agent/internal/mq/message"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPDeliverySource interface {
	Consume(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

type RabbitMQConsumer struct {
	source  AMQPDeliverySource
	queue   string
	handler Handler
}

func NewRabbitMQConsumer(source AMQPDeliverySource, queue string, handler Handler) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		source:  source,
		queue:   queue,
		handler: handler,
	}
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	deliveries, err := c.source.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}
			c.handleDelivery(ctx, delivery)
		}
	}
}

func (c *RabbitMQConsumer) handleDelivery(ctx context.Context, delivery amqp.Delivery) {
	var msg message.TaskMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		_ = delivery.Nack(false, false)
		return
	}

	if err := c.handler.Handle(ctx, msg); err != nil {
		_ = delivery.Nack(false, true)
		return
	}

	_ = delivery.Ack(false)
}
