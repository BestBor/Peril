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

const (
	connString = "amqp://guest:guest@localhost:5672/"
)

func PublishGameLog(ch *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			CurrentTime: time.Now(),
			Message:     msg,
			Username:    username,
		},
	)
}

func main() {
	fmt.Println("Starting Peril client...")
	rabbitConn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("unable to create channel")
	}
	defer rabbitChan.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		rabbitConn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("unable to subscribe to pause")
	}

	err = pubsub.SubscribeJSON(
		rabbitConn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, rabbitChan),
	)
	if err != nil {
		log.Fatalf("unable to subscribe to army_moves")
	}

	err = pubsub.SubscribeJSON(
		rabbitConn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gameState, rabbitChan),
	)
	if err != nil {
		log.Fatalf("unable to subscribe to war_moves")
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err = gameState.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			am, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				rabbitChan,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				am,
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(am.Units), am.ToLocation)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			// fmt.Println("Spamming not allowed yet!")
			if len(words) > 1 {
				quantity, err := strconv.Atoi(words[1])
				if err != nil {
					fmt.Printf("error: %s\n", err)
					continue
				}
				for range quantity {
					mlm := gamelogic.GetMaliciousLog()
					err := PublishGameLog(rabbitChan, username, mlm)
					if err != nil {
						fmt.Printf("error: %s\n", err)
						continue
					}

				}
			} else {
				println("error: command incomplete")
				continue
			}
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("unknown command")
		}
	}

	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
}
