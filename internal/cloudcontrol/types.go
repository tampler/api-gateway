package cloudcontrol

import (
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"go.uber.org/zap"
)

type NatsConfig struct {
	Timeout int
	Server  string
	Topic   string
}

type APIMessage struct {
	Service  string
	Resource string
	Action   string
	Cfg      NatsConfig
}

// APIServer - top level execution engine
type APIServer struct {
	zl  *zap.Logger
	cfg *config.AppConfig
}
