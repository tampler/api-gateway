package apiserver

import (
	"context"
	"fmt"

	aj "github.com/choria-io/asyncjobs"
	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
)

type MyContext struct {
	echo.Context
	pub *Publisher
}

func MakeMyContext(c echo.Context, pub *Publisher) *MyContext {
	return &MyContext{c, pub}
}

func (c *MyContext) Foo() {
	println("foo")
}

type BusEvent struct {
	data []byte
}

type Subscriber interface {
	Notify(BusEvent)
}

type Publisher struct {
	sub  Subscriber
	pong QueueManager
}

func MakePublisher(m QueueManager) Publisher {
	return Publisher{pong: m}
}

func (p *Publisher) AddHandlers() error {

	err := p.pong.router.HandleFunc(topic, func(ctx context.Context, _ aj.Logger, t *aj.Task) (interface{}, error) {

		data, err := decodeJSONBytes(t.Payload)
		if err != nil {
			return nil, err
		}

		fmt.Printf("*** PONG handler with PLOAD %v\n", string(data))

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
	ID      int
	Message []byte
	done    chan bool
}

func (p *BusObserver) Notify(ev BusEvent) {
	fmt.Printf(" *** NOTIFY %v received \n", string(ev.data))
	p.Message = ev.data
	p.done <- true
}
