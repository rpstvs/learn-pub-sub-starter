package pubsub

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int
type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()

	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, errors.New("couldnt open channel on conn")
	}

	queue, err := ch.QueueDeclare(queueName, simpleQueueType == int(SimpleQueueDurable), simpleQueueType == int(SimpleQueueTransient), simpleQueueType == int(SimpleQueueTransient), false, nil)

	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, errors.New("couldnt create queue")
	}

	err = ch.QueueBind(queue.Name, key, exchange, false, nil)

	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	return ch, queue, nil

}
