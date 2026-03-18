package main

import (
	"fmt"
	"os"
	"os/signal"

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
		return
	}
	defer rabbitConn.Close()

	channel, err := rabbitConn.Channel()
	if err != nil {
		return
	}

	pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	fmt.Println("Peril connection to RabbitMQ successful!")

	// Sign Channel
	signalChan := make(chan os.Signal, 1)

	// Listen to Ctrl + C (SIGINT) & SIGTERM
	signal.Notify(signalChan, os.Interrupt)

	// block until recieved signal
	<-signalChan

	fmt.Println("Shutting down connection...")
}
