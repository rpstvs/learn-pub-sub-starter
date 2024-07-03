package main

import (
	"fmt"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connString := "amqp://guest:guest@localhost:5672/"
	cenas, err := amqp.Dial(connString)
	defer cenas.Close()
	if err != nil {
		fmt.Println("connection not done")
	}

	fmt.Println("Connection Successful")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	shuttingDown := <-signalChan
	if shuttingDown != nil {
		fmt.Println("Program is shutting down.")
	}

}
