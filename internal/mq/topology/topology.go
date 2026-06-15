package topology

import (
	"campus-agent/internal/config"
	"campus-agent/internal/mq/message"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Channel interface {
	ExchangeDeclare(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp.Table) error
	QueueDeclare(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name string, key string, exchange string, noWait bool, args amqp.Table) error
}

func DeclareTaskExecution(channel Channel, cfg config.RabbitMQConfig) error {
	exchange := cfg.Exchange
	if exchange == "" {
		exchange = message.ExchangeCampusAgent
	}

	queue := cfg.TaskExecuteQueue
	if queue == "" {
		queue = message.QueueTaskExecute
	}

	if err := channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	if _, err := channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		return err
	}

	return channel.QueueBind(queue, queue, exchange, false, nil)
}
