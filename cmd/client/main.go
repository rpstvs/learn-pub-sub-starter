package main

import (
	"fmt"
	"log"

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

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+"."+"*", int(pubsub.SimpleQueueDurable))

	if err != nil {
		log.Fatal("couldnt declare queue for gamelog")
	}

	fmt.Printf("Queue %v declared and bound! \n", queue.Name)

	gs := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, int(pubsub.SimpleQueueTransient), handlerPause(gs))

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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
