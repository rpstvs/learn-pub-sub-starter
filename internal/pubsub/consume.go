package pubsub

import (
	"encoding/json"
	"errors"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int
type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)

	if err != nil {
		log.Println("ch and queue binded")
		return err
	}

	delivery, err := ch.Consume(queue.Name, "", false, false, false, false, nil)

	if err != nil {
		log.Println("error creating consumer")
		return err
	}

	go func() {
		defer ch.Close()
		for msg := range delivery {
			var data T

			err = json.Unmarshal(msg.Body, &data)

			if err != nil {
				log.Println("error decoding message body")
			}
			acknowledge := handler(data)

			switch acknowledge {
			case Ack:
				msg.Ack(false)
				log.Printf("Handler returned %d \n", acknowledge)
			case NackDiscard:
				msg.Nack(false, false)
				log.Printf("Handler returned %d \n", acknowledge)
			case NackRequeue:
				msg.Nack(false, true)
				log.Printf("Handler returned %d \n", acknowledge)
			}

		}
	}()

	return nil
}
