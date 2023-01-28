package util

import (
	"log"

	"github.com/choria-io/asyncjobs"
	"github.com/nats-io/nats.go"
)

func AddTaskToQueue() {
	client, err := asyncjobs.NewClient(asyncjobs.NatsConn(&nats.Conn{}), asyncjobs.BindWorkQueue("link"))
	if err != nil {
		log.Fatalf("failed to created new asyncjobs client with error: %v", err)
	}

	log.Println(client)
}
