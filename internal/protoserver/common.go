package protoserver

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
)

type storageServer struct {
	kv nats.KeyValue
}

func MakeStorageServer(nc *nats.Conn, cfg config.AppConfig) (storageServer, error) {

	var serv storageServer

	// Setup a JetStream
	js, err := nc.JetStream()
	if err != nil {
		return serv, fmt.Errorf("NATS JetStream failed %s \n", err.Error())
	}

	kv, err := js.KeyValue(cfg.Sdk.Bucket)
	if err != nil {
		return serv, fmt.Errorf("NATS KeyValue failed %s \n", err.Error())
	}

	serv.kv = kv

	return serv, nil
}
