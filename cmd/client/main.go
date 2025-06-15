package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	username, _ := gamelogic.ClientWelcome()

	_, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, int(pubsub.SimpleQueueTransient))

	if err != nil {
		log.Fatal("couldnt declare or bind")
	}

	fmt.Printf("Queue %v declared and bound! \n", queue.Name)

	gs := gamelogic.NewGameState(username)
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	fmt.Println("Rabbit mq closed conn")

	for {
		commands := gamelogic.GetInput()

		if len(commands) == 0 {
			continue
		}

		switch commands[0] {
		case "spawn":
			err = gs.CommandSpawn(commands)

			if err != nil {
				log.Println(err)
			}
		case "move":
			_, err = gs.CommandMove(commands)

			if err != nil {
				log.Println(err)
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			log.Println("Spamming not allowed")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Println("command not valid")
		}

	}
}
