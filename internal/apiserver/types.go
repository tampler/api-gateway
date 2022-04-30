package apiserver

import (
	"context"

	aj "github.com/choria-io/asyncjobs"
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"go.uber.org/zap"
)

// QueueManager - an async queue job manager
type QueueManager struct {
	client *aj.Client
	router *aj.Mux
}

// Run - runs the client and returns an erorr in a channel
func (m QueueManager) Run(ctx context.Context, c chan error) {
	err := m.client.Run(ctx, m.router)
	c <- err
}

// APIServer - top level execution engine
type APIServer struct {
	zl   *zap.Logger
	cfg  *config.AppConfig
	ping QueueManager
	pong QueueManager
}

// MakeAPIServer - APIServer factory
func MakeAPIServer(c *config.AppConfig, z *zap.Logger, pingC, pongC *aj.Client, pingR, pongR *aj.Mux) *APIServer {
	srv := APIServer{
		zl:   z,
		cfg:  c,
		ping: QueueManager{client: pingC, router: pingR},
		pong: QueueManager{client: pongC, router: pongR},
	}
	return &srv
}

// APICommand - API command
type APIRequestCommand struct {
	JobID    string
	Service  string
	Resource string
	Action   string
	Params   []string
}

// APIMessage - message to be processed by SDK
type APIRequestMessage struct {
	Cmd APIRequestCommand
}

// APIResponseMessage - final JSON output from SDK
type APIResponseMessage struct {
	Service string `json:"service"`
	Api     string `json:"api"`
	Data    []byte `json:"data"`
}
