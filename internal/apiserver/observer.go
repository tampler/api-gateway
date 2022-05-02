package apiserver

import (
	"context"
	"fmt"

	aj "github.com/choria-io/asyncjobs"
	"github.com/labstack/echo/v4"
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
	subs []Subscriber
	pong QueueManager
}

func MakePublisher(m QueueManager) Publisher {
	return Publisher{pong: m}
}

func (p *Publisher) AddHandlers() error {

	err := p.pong.router.HandleFunc(topic, func(ctx context.Context, _ aj.Logger, t *aj.Task) (interface{}, error) {
		fmt.Printf("*** PONG handler for task %v\n", t.ID)

		data, err := decodeJSONBytes(t.Payload)
		if err != nil {
			return nil, err
		}

		fmt.Printf(">>>> Started NOTIFY... \n")

		p.NotifyObservers(BusEvent{
			data: data,
		})

		fmt.Printf(">>>> Exit NOTIFY... \n")

		return nil, nil
	})
	if err != nil {
		return err
	}

	// Execute PONG queue
	go p.pong.Run(context.Background())

	return nil
}

func (p *Publisher) AddSubscriber(o Subscriber) {
	p.subs = append(p.subs, o)
}

func (p *Publisher) RemoveObserver(o Subscriber) {
	var indexToRemove int
	for i, observer := range p.subs {
		if observer == o {
			indexToRemove = i
			break
		}
	}
	p.subs = append(
		p.subs[:indexToRemove],
		p.subs[indexToRemove+1:]...,
	)
}

func (p *Publisher) NotifyReciever(id string, e BusEvent) {
	for _, observer := range p.subs {
		observer.Notify(e)
	}
}

func (p *Publisher) NotifyObservers(ev BusEvent) {
	for _, observer := range p.subs {
		observer.Notify(ev)
	}
}

type TestObserver struct {
	ID      int
	Message []byte
	done    chan bool
}

func (p *TestObserver) Notify(ev BusEvent) {
	fmt.Printf("Obderver %d: message '%s' received \n", p.ID, ev.data)
	p.Message = ev.data
	p.done <- true
}
