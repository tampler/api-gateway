package worker

import (
	"context"

	aj "github.com/choria-io/asyncjobs"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"go.uber.org/zap"
)

// QueueManager - an async queue job manager
type QueueManager struct {
	Client *aj.Client
	Router *aj.Mux
}

// MakeQueueManager - factory for QueueManager
func MakeQueueManager(c *aj.Client, r *aj.Mux) QueueManager {
	return QueueManager{Client: c, Router: r}
}

// Run - runs the client and returns an erorr in a channel
func (m QueueManager) Run(ctx context.Context) error {
	return m.Client.Run(ctx, m.Router)
}

// APIServer - top level execution engine
type APIServer struct {
	zl   *zap.SugaredLogger
	cfg  *config.AppConfig
	Ping QueueManager
	Pong QueueManager
}

// MakeAPIServer - APIServer factory
func MakeAPIServer(c *config.AppConfig, z *zap.SugaredLogger, ping, pong QueueManager) *APIServer {
	srv := APIServer{
		zl:   z,
		cfg:  c,
		Ping: ping,
		Pong: pong,
	}
	return &srv
}
