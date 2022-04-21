package apiserver

import (
	"github.com/nats-io/nats.go"
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"go.uber.org/zap"
)

// APIServer - top level execution engine
type APIServer struct {
	nats *nats.Conn
	zl   *zap.Logger
	cfg  *config.AppConfig
}

// NatsConfig - Nats configuration
type NatsConfig struct {
	Timeout int
	Server  string
	Topic   string
}

// APICommand - API command
type APIRequestCommand struct {
	Service  string
	Resource string
	Action   string
	Params   []string
}

// APIMessage - message to be processed by SDK
type APIRequestMessage struct {
	Cfg NatsConfig
	Cmd APIRequestCommand
}

// APIResponseMessage - final JSON output from SDK
type APIResponseMessage struct {
	Service string `json:"service"`
	Api     string `json:"api"`
	Data    []byte `json:"data"`
}
