package apiserver

import (
	"context"

	aj "github.com/choria-io/asyncjobs"
	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"go.uber.org/zap"
)

// MyContext - custom echo context
type MyContext struct {
	echo.Context
	zl  *zap.SugaredLogger
	pub *Publisher
}

// MakeMyContext - factory to create a context
func MakeMyContext(c echo.Context, pub *Publisher, zl *zap.SugaredLogger) *MyContext {
	return &MyContext{
		c, zl, pub,
	}
}

// QueueManager - an async queue job manager
type QueueManager struct {
	client *aj.Client
	router *aj.Mux
}

// MakeQueueManager - factory for QueueManager
func MakeQueueManager(c *aj.Client, r *aj.Mux) QueueManager {
	return QueueManager{client: c, router: r}
}

// Run - runs the client and returns an erorr in a channel
func (m QueueManager) Run(ctx context.Context) error {
	return m.client.Run(ctx, m.router)
}

// APIServer - top level execution engine
type APIServer struct {
	zl   *zap.SugaredLogger
	cfg  *config.AppConfig
	ping QueueManager
	pong QueueManager
}

// MakeAPIServer - APIServer factory
func MakeAPIServer(c *config.AppConfig, z *zap.SugaredLogger, ping, pong QueueManager) *APIServer {
	srv := APIServer{
		zl:   z,
		cfg:  c,
		ping: ping,
		pong: pong,
	}
	return &srv
}

// APICommand - cloud control command
type APICommand struct {
	Service  string
	Resource string
	Action   string
	Params   []string
}

// APIRequest - top level API request
type APIRequest struct {
	JobID string
	Cmd   APICommand
}

// APIResponseMessage - final JSON output from SDK
type APIResponseMessage struct {
	Service string `json:"service"`
	Api     string `json:"api"`
	Data    []byte `json:"data"`
}
