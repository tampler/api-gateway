package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	aj "github.com/choria-io/asyncjobs"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type SubMap = map[uuid.UUID]Subscriber

type BusEvent struct {
	data []byte
	err  string
}

type Subscriber interface {
	Notify(BusEvent)
}

type Publisher struct {
	mutex sync.RWMutex
	sub   SubMap
	pong  QueueManager
	zl    *zap.SugaredLogger
}

func MakePublisher(m QueueManager, zl *zap.SugaredLogger, sm SubMap) Publisher {
	return Publisher{pong: m, zl: zl, sub: sm}
}

func (p *Publisher) AddHandlers(topic string) error {

	err := p.pong.router.HandleFunc(topic, func(ctx context.Context, _ aj.Logger, t *aj.Task) (interface{}, error) {

		var resp APIResponse

		if err := json.Unmarshal(t.Payload, &resp); err != nil {
			p.zl.Error(err)
			return nil, err
		}

		p.NotifyObserver(resp.JobID, BusEvent{data: resp.Data, err: resp.Err})

		return nil, nil
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
	// p.zl.Debugf("Adding observer: %v", id)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.sub[id] = sub
}

func (p *Publisher) RemoveObserver(id uuid.UUID) {
	// p.zl.Debugf("Removing observer: %v", id)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.sub[id] = nil
}

func (p *Publisher) NotifyObserver(id uuid.UUID, e BusEvent) {
	// p.zl.Debugf("Notifying observer: %v", id)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.sub[id].Notify(e)
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
	bo.data = ev.data
	bo.done <- true
}
