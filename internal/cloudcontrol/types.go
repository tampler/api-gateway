package cloudcontrol

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
type APICommand struct {
	Service  string
	Resource string
	Action   string
	Params   []string
}

// APIMessage - message to be processed by SDK
type APIMessage struct {
	Cfg NatsConfig
	Cmd APICommand
}
