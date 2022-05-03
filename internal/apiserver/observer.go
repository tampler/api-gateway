package apiserver

import (
	"context"
	"fmt"

	aj "github.com/choria-io/asyncjobs"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
	"go.uber.org/zap"
)

type BusEvent struct {
	data []byte
}

type Subscriber interface {
	Notify(BusEvent)
}

type Publisher struct {
	sub  Subscriber
	pong QueueManager
	zl   *zap.SugaredLogger
}

func MakePublisher(m QueueManager, zl *zap.SugaredLogger) Publisher {
	return Publisher{pong: m, zl: zl}
}

func (p *Publisher) AddHandlers() error {

	err := p.pong.router.HandleFunc(topic, func(ctx context.Context, _ aj.Logger, t *aj.Task) (interface{}, error) {

		data, err := decodeJSONBytes(t.Payload)
		if err != nil {
			return nil, err
		}

		p.zl.Infof("*** PONG handler with PLOAD %v\n", string(data))

		p.NotifyObservers(BusEvent{
			data: data,
		})

		return nil, nil
	})
	if err != nil {
		return fail.Error500(fmt.Sprintf("PONG Handler: %v\n", err.Error()))
	}

	// Execute PONG queue
	go p.pong.Run(context.Background())

	return nil
}

func (p *Publisher) AddSubscriber(o Subscriber) {
	p.sub = o
}

func (p *Publisher) RemoveObserver(o Subscriber) {
	p.sub = nil
}

func (p *Publisher) NotifyReciever(id string, e BusEvent) {
	p.sub.Notify(e)
}

func (p *Publisher) NotifyObservers(ev BusEvent) {
	p.sub.Notify(ev)

}

type BusObserver struct {
	id   int
	data []byte
	zl   *zap.SugaredLogger
	done chan bool
}

func MakeBusObserver(id int, data []byte, zl *zap.SugaredLogger, done chan bool) BusObserver {
	return BusObserver{id: id, data: data, zl: zl, done: done}
}

func (bo *BusObserver) Notify(ev BusEvent) {
	bo.zl.Infof(" *** NOTIFY %v received \n", string(ev.data))
	bo.data = ev.data
	bo.done <- true
}
