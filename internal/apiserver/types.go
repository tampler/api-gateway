package apiserver

import (
	"context"
	"fmt"

	aj "github.com/choria-io/asyncjobs"
	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"go.uber.org/zap"
)

type handlerConfig struct {
	ctx      echo.Context
	data     []byte
	service  string
	resource string
}

type HandlerFunctor = func(handlerConfig) error

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

func (m QueueManager) SetupHandler(cfg handlerConfig, method HandlerFunctor) error {

	err := m.router.HandleFunc(topic, func(_ context.Context, _ aj.Logger, t *aj.Task) (interface{}, error) {

		fmt.Printf("*** Processing PONG task ID: %s\n", t.ID)

		err := method(cfg)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	return err
}

// APIServer - top level execution engine
type APIServer struct {
	zl   *zap.Logger
	cfg  *config.AppConfig
	ping QueueManager
	pong QueueManager
}

// MakeAPIServer - APIServer factory
func MakeAPIServer(c *config.AppConfig, z *zap.Logger, ping, pong QueueManager) *APIServer {
	srv := APIServer{
		zl:   z,
		cfg:  c,
		ping: ping,
		pong: pong,
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
