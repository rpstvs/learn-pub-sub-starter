package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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

	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gs))
	if err != nil {
		log.Fatal("couldnt subscribe to pause queue")
	}
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+gs.GetUsername(), routing.ArmyMovesPrefix+".*", pubsub.SimpleQueueTransient, handlerMoves(gs, publishCh))

	if err != nil {
		log.Fatal("couldnt subscribe to army move queue")
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.SimpleQueueDurable, handleWar(gs, publishCh))

	if err != nil {
		log.Fatal("couldnt subscribe to War")
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
			if len(commands) < 2 {
				continue
			}
			n, err := strconv.Atoi(commands[1])

			if err != nil {
				log.Println("number couldnt be converted")
				continue
			}
			str := gamelogic.GetMaliciousLog()

			for i := 0; i < n; i++ {
				err := pubsub.PublishGob(publishCh, routing.ExchangePerilTopic, routing.GameLogSlug+"."+gs.GetUsername(), routing.GameLog{
					Username:    gs.GetUsername(),
					Message:     str,
					CurrentTime: time.Now(),
				})
				if err != nil {
					fmt.Println("Couldnt publish malicious log")
					continue
				}
			}

			log.Println("Spamming not allowed")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Println("command not valid")
		}

	}
}
