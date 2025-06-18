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

	publishCh, err := conn.Channel()

	if err != nil {
		log.Fatal("could not create channel")
	}
	username, _ := gamelogic.ClientWelcome()

	fmt.Printf("Queue %v declared and bound! \n")

	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, int(pubsub.SimpleQueueTransient), handlerPause(gs))
	if err != nil {
		log.Fatal("couldnt subscribe to pause queue")
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, routing.ArmyMovesPrefix+".*", int(pubsub.SimpleQueueTransient), handlerMoves(gs))

	if err != nil {
		log.Fatal("couldnt subscribe to army move queue")
	}
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
			gl, err := gs.CommandMove(commands)

			if err != nil {
				log.Println(err)
			}
			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+gl.Player.Username, gl)
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
