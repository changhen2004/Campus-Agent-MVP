package consumer

import (
	"context"

	"campus-agent/internal/mq/message"
)

type Handler interface {
	Handle(ctx context.Context, msg message.TaskMessage) error
}

type TaskConsumer struct {
	handler Handler
}

func NewTaskConsumer(handler Handler) *TaskConsumer {
	return &TaskConsumer{handler: handler}
}

func (c *TaskConsumer) Consume(ctx context.Context, msg message.TaskMessage) error {
	if c.handler == nil {
		return nil
	}
	return c.handler.Handle(ctx, msg)
}
