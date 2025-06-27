package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
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
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()

	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, errors.New("couldnt open channel on conn")
	}

	queue, err := ch.QueueDeclare(queueName, queueType == SimpleQueueDurable, queueType == SimpleQueueTransient, queueType == SimpleQueueTransient, false, amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	})

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
	simpleQueueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, func(b []byte) (T, error) {
		var out T
		err := json.Unmarshal(b, &out)
		return out, err
	})
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	return subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		func(data []byte) (T, error) {
			buffer := bytes.NewBuffer(data)
			decoder := gob.NewDecoder(buffer)
			var target T
			err := decoder.Decode(&target)
			return target, err
		},
	)
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
	unmarshaller func([]byte) (T, error),
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
			data, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
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
