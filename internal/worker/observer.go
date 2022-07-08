package worker

import (
	"context"
	"encoding/json"
	"fmt"

	aj "github.com/choria-io/asyncjobs"
	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
	"go.uber.org/zap"
)

// AddHandlers - add AJC queue handlers for a given topic
func (p *Publisher) AddHandlers(topic string) error {

	err := p.Pong.Router.HandleFunc(topic, func(ctx context.Context, _ aj.Logger, t *aj.Task) (interface{}, error) {

		var resp cc.APIResponse

		if err := json.Unmarshal(t.Payload, &resp); err != nil {
			p.Zl.Error(err)
			return nil, err
		}

		id, err := uuid.FromBytes(resp.JobID)
		if err != nil {
			p.Zl.Error(err)
			return nil, err
		}

		return nil, p.NotifyObserver(id, BusEvent{Data: resp.Data, Err: resp.Err})
	})

	if err != nil {
		p.Zl.Error(err)
		return fail.Error500(fmt.Sprintf("PONG Handler: %v", err.Error()))
	}

	// Execute PONG queue
	go p.Pong.Run(context.Background())

	return nil
}

func (p *Publisher) AddObserver(id uuid.UUID, sub Subscriber) {
	// p.zl.Debugf("Adding observer: %s", id)
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.Sub[id] = sub
}

func (p *Publisher) RemoveObserver(id uuid.UUID) {
	// p.zl.Debugf("Removing observer: %s", id)
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	delete(p.Sub, id)
}

func (p *Publisher) NotifyObserver(id uuid.UUID, e BusEvent) error {
	// p.zl.Debugf("Notifying observer: %s", id)
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	// Safe access to map
	if _, ok := p.Sub[id]; !ok {
		msg := fmt.Sprintf("Observer %s not found", id)
		p.Zl.Error(msg)
		return fail.Error500(msg)
	}

	p.Sub[id].Notify(e)
	return nil
}

// BusObserver - AJC async listener
type BusObserver struct {
	zl   *zap.SugaredLogger
	id   uuid.UUID
	Data []byte
	Err  string
	Done chan bool
}

// MakeBusObserver - factory for Bus observer
func MakeBusObserver(id uuid.UUID, zl *zap.SugaredLogger, done chan bool) BusObserver {
	return BusObserver{id: id, zl: zl, Done: done}
}

// Notify - notification with unblocking for listeners
func (bo *BusObserver) Notify(ev BusEvent) {
	bo.Err = ev.Err

	// pass an empty buffer to avoid exceptions for empty buffer response from SDK
	if ev.Data != nil {
		bo.Data = ev.Data
	} else {
		ev.Data = []byte{}
	}

	// Confirm process finish
	bo.Done <- true
}
