package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	connString = "amqp://guest:guest@localhost:5672/"
)

func main() {
	fmt.Println("Starting Peril server...")

	rabbitConn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	publishCh, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("unable to create channel: %v", err)
	}

	pubsub.DeclareAndBind(
		rabbitConn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
	)

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("Sending pause message...")
			pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		case "resume":
			fmt.Println("Sending resume message...")
			pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		case "quit":
			fmt.Println("Exiting system...")
			return
		default:
			fmt.Println("Command not understood")
		}
	}

	// Sign Channel
	// signalChan := make(chan os.Signal, 1)

	// // Listen to Ctrl + C (SIGINT) & SIGTERM
	// signal.Notify(signalChan, os.Interrupt)

	// // block until recieved signal
	// <-signalChan

	// fmt.Println("Shutting down connection...")
}
