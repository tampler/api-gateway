package cloudcontrol

import (
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"go.uber.org/zap"
)

type NatsConfig struct {
	timeout int
	server  string
	topic   string
}

type Command struct {
	service  string
	resource string
	action   string
	cfg      NatsConfig
}

// APIServer - top level execution engine
type APIServer struct {
	zl  *zap.Logger
	cfg *config.AppConfig
}
