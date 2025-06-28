package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerMoves(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		mv := gs.HandleMove(move)

		switch mv {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.GetUsername(), gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			})

			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		}
		fmt.Println("Move not known")
		return pubsub.NackDiscard
	}
}

func handleWar(gs *gamelogic.GameState, ch *amqp.Channel) func(war gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(war gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		warOut, winner, loser := gs.HandleWar(war)
		switch warOut {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			log := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+war.Attacker.Username, routing.GameLog{
				CurrentTime: time.Now(),
				Message:     log,
				Username:    gs.GetUsername(),
			})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:

			log := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+war.Attacker.Username, routing.GameLog{
				CurrentTime: time.Now(),
				Message:     log,
				Username:    gs.GetUsername(),
			})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			log := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			err := pubsub.PublishGob(ch, routing.GameLogSlug, routing.GameLogSlug+"."+war.Attacker.Username, routing.GameLog{
				CurrentTime: time.Now(),
				Message:     log,
				Username:    gs.GetUsername(),
			})
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		return pubsub.NackDiscard
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}
