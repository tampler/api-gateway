package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/nats-io/nats.go"
)

const (
	topic    = "MyTopic"
	NATS_URL = "nats://hyp:4222"
)

func main() {

	// Create server connection
	nc, err := nats.Connect(NATS_URL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS server %v", NATS_URL)
	}

	log.Println("Connected to " + NATS_URL)

	// Subscribe to subject
	nc.Subscribe(topic, func(msg *nats.Msg) {
		fmt.Printf("*** Sub: got message: %v \n", string(msg.Data))
		nc.Publish(msg.Reply, []byte("I'm here"))
	})

	// Keep the connection alive
	runtime.Goexit()
}
