package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	aj "github.com/choria-io/asyncjobs"
	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
	"go.uber.org/zap"
)

// BusEvent - bus event structure
type BusEvent struct {
	data []byte
	err  string
}

// Subscriber - bus subscriber
type Subscriber interface {
	Notify(BusEvent)
}

// Publisher - bus publisher
type Publisher struct {
	mutex sync.RWMutex
	sub   SubMap
	pong  worker.QueueManager
	zl    *zap.SugaredLogger
}

// MakePublisher - factory for Publisher
func MakePublisher(m worker.QueueManager, zl *zap.SugaredLogger, sm SubMap) Publisher {
	return Publisher{pong: m, zl: zl, sub: sm}
}

// AddHandlers - add AJC queue handlers for a given topic
func (p *Publisher) AddHandlers(topic string) error {

	err := p.pong.Router.HandleFunc(topic, func(ctx context.Context, _ aj.Logger, t *aj.Task) (interface{}, error) {

		var resp APIResponse

		if err := json.Unmarshal(t.Payload, &resp); err != nil {
			p.zl.Error(err)
			return nil, err
		}

		return nil, p.NotifyObserver(resp.JobID, BusEvent{data: resp.Data, err: resp.Err})
	})

	if err != nil {
		p.zl.Error(err)
		return fail.Error500(fmt.Sprintf("PONG Handler: %v", err.Error()))
	}

	// Execute PONG queue
	go p.pong.Run(context.Background())

	return nil
}

func (p *Publisher) AddObserver(id uuid.UUID, sub Subscriber) {
	// p.zl.Debugf("Adding observer: %s", id)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.sub[id] = sub
}

func (p *Publisher) RemoveObserver(id uuid.UUID) {
	// p.zl.Debugf("Removing observer: %s", id)
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.sub, id)
}

func (p *Publisher) NotifyObserver(id uuid.UUID, e BusEvent) error {
	// p.zl.Debugf("Notifying observer: %s", id)
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Safe access to map
	if _, ok := p.sub[id]; !ok {
		msg := fmt.Sprintf("Observer %s not found", id)
		p.zl.Error(msg)
		return fail.Error500(msg)
	}

	p.sub[id].Notify(e)
	return nil
}

// BusObserver - AJC async listener
type BusObserver struct {
	zl   *zap.SugaredLogger
	id   uuid.UUID
	data []byte
	err  string
	done chan bool
}

// MakeBusObserver - factory for Bus observer
func MakeBusObserver(id uuid.UUID, zl *zap.SugaredLogger, done chan bool) BusObserver {
	return BusObserver{id: id, zl: zl, done: done}
}

// Notify - notification with unblocking for listeners
func (bo *BusObserver) Notify(ev BusEvent) {
	bo.err = ev.err

	// pass an empty buffer to avoid exceptions for empty buffer response from SDK
	if ev.data != nil {
		bo.data = ev.data
	} else {
		ev.data = []byte{}
	}

	// Confirm process finish
	bo.done <- true
}
