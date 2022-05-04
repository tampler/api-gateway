package apiserver

import (
	"context"
	"encoding/json"
	"fmt"

	aj "github.com/choria-io/asyncjobs"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type SubMap = map[uuid.UUID]Subscriber

type BusEvent struct {
	data []byte
}

type Subscriber interface {
	Notify(BusEvent)
}

type Publisher struct {
	sub  SubMap
	pong QueueManager
	zl   *zap.SugaredLogger
}

func MakePublisher(m QueueManager, zl *zap.SugaredLogger, sm SubMap) Publisher {
	return Publisher{pong: m, zl: zl, sub: sm}
}

func (p *Publisher) AddHandlers() error {

	err := p.pong.router.HandleFunc(topic, func(ctx context.Context, log aj.Logger, t *aj.Task) (interface{}, error) {

		var resp APIResponse

		if err := json.Unmarshal(t.Payload, &resp); err != nil {
			return nil, aj.ErrTerminateTask
		}

		data, err := decodeJSONBytes(resp.Data)
		if err != nil {
			log.Errorf("PONG failed to decode a JSON payload")
			return nil, aj.ErrTerminateTask
		}

		p.NotifyObserver(resp.JobID, BusEvent{data: data})

		return nil, nil
	})
	if err != nil {
		return fail.Error500(fmt.Sprintf("PONG Handler: %v\n", err.Error()))
	}

	// Execute PONG queue
	go p.pong.Run(context.Background())

	return nil
}

func (p *Publisher) AddObserver(id uuid.UUID, sub Subscriber) {
	p.zl.Debugf("Adding observer: %v", id)
	p.sub[id] = sub
}

func (p *Publisher) RemoveObserver(id uuid.UUID) {
	p.zl.Debugf("Removing observer: %v", id)
	p.sub[id] = nil
}

func (p *Publisher) NotifyObserver(id uuid.UUID, e BusEvent) {
	p.zl.Debugf("Notifying observer: %v", id)
	p.sub[id].Notify(e)
}

// BusObserver - AJC async listener
type BusObserver struct {
	id   uuid.UUID
	data []byte
	zl   *zap.SugaredLogger
	done chan bool
}

// MakeBusObserver - factory for Bus observer
func MakeBusObserver(id uuid.UUID, data []byte, zl *zap.SugaredLogger, done chan bool) BusObserver {
	return BusObserver{id: id, data: data, zl: zl, done: done}
}

// Notify - notification with unblocking for listeners
func (bo *BusObserver) Notify(ev BusEvent) {
	bo.data = ev.data
	bo.done <- true
}
