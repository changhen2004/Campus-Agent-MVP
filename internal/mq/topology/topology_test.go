package topology

import (
	"testing"

	"campus-agent/internal/config"
	"campus-agent/internal/mq/message"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDeclareTaskExecutionTopology(t *testing.T) {
	t.Parallel()

	channel := &fakeTopologyChannel{}
	cfg := config.RabbitMQConfig{
		Exchange:         message.ExchangeCampusAgent,
		TaskExecuteQueue: message.QueueTaskExecute,
	}

	if err := DeclareTaskExecution(channel, cfg); err != nil {
		t.Fatalf("DeclareTaskExecution returned error: %v", err)
	}

	if channel.exchangeName != message.ExchangeCampusAgent {
		t.Fatalf("exchange mismatch: got %q want %q", channel.exchangeName, message.ExchangeCampusAgent)
	}

	if channel.exchangeKind != "direct" {
		t.Fatalf("exchange kind mismatch: got %q want %q", channel.exchangeKind, "direct")
	}

	if !channel.exchangeDurable {
		t.Fatal("exchange should be durable")
	}

	if channel.queueName != message.QueueTaskExecute {
		t.Fatalf("queue mismatch: got %q want %q", channel.queueName, message.QueueTaskExecute)
	}

	if !channel.queueDurable {
		t.Fatal("queue should be durable")
	}

	if channel.bindingQueue != message.QueueTaskExecute {
		t.Fatalf("binding queue mismatch: got %q want %q", channel.bindingQueue, message.QueueTaskExecute)
	}

	if channel.bindingKey != message.QueueTaskExecute {
		t.Fatalf("binding key mismatch: got %q want %q", channel.bindingKey, message.QueueTaskExecute)
	}

	if channel.bindingExchange != message.ExchangeCampusAgent {
		t.Fatalf("binding exchange mismatch: got %q want %q", channel.bindingExchange, message.ExchangeCampusAgent)
	}
}

type fakeTopologyChannel struct {
	exchangeName    string
	exchangeKind    string
	exchangeDurable bool
	queueName       string
	queueDurable    bool
	bindingQueue    string
	bindingKey      string
	bindingExchange string
}

func (c *fakeTopologyChannel) ExchangeDeclare(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp.Table) error {
	c.exchangeName = name
	c.exchangeKind = kind
	c.exchangeDurable = durable
	return nil
}

func (c *fakeTopologyChannel) QueueDeclare(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (amqp.Queue, error) {
	c.queueName = name
	c.queueDurable = durable
	return amqp.Queue{Name: name}, nil
}

func (c *fakeTopologyChannel) QueueBind(name string, key string, exchange string, noWait bool, args amqp.Table) error {
	c.bindingQueue = name
	c.bindingKey = key
	c.bindingExchange = exchange
	return nil
}
