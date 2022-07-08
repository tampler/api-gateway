package worker

import (
	"context"
	"sync"

	aj "github.com/choria-io/asyncjobs"
	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"go.uber.org/zap"
)

// SubMap - event subscriber map
type SubMap = map[uuid.UUID]Subscriber

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
	Zl   *zap.SugaredLogger
	Cfg  *config.AppConfig
	Ping QueueManager
	Pong QueueManager
}

// MakeAPIServer - APIServer factory
func MakeAPIServer(c *config.AppConfig, z *zap.SugaredLogger, ping, pong QueueManager) *APIServer {
	srv := APIServer{
		Zl:   z,
		Cfg:  c,
		Ping: ping,
		Pong: pong,
	}
	return &srv
}

// Publisher - bus publisher
type Publisher struct {
	Mutex sync.RWMutex
	Sub   SubMap
	Pong  QueueManager
	Zl    *zap.SugaredLogger
}

// MakePublisher - factory for Publisher
func MakePublisher(m QueueManager, zl *zap.SugaredLogger, sm SubMap) Publisher {
	return Publisher{Pong: m, Zl: zl, Sub: sm}
}

// BusEvent - bus event structure
type BusEvent struct {
	Data []byte
	Err  string
}

// Subscriber - bus subscriber
type Subscriber interface {
	Notify(BusEvent)
}
