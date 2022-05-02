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

func (c *MyContext) Foo() {
	println("foo")
}

type Subscriber interface {
	Notify(string)
}

type Publisher struct {
	ObserverList []Subscriber
	pong         QueueManager
}

func (p *Publisher) AddHandlers() error {

	err := p.pong.router.HandleFunc(topic, func(ctx context.Context, _ aj.Logger, t *aj.Task) (interface{}, error) {
		fmt.Printf("*** PONG handler for task %v\n", t.ID)

		return nil, nil
	})
	if err != nil {
		return err
	}

	ch := make(chan error)
	go p.pong.Run(context.Background(), ch)

	return nil
}

func (p *Publisher) AddSubscriber(o Subscriber) {
	p.ObserverList = append(p.ObserverList, o)
}

func (p *Publisher) RemoveObserver(o Subscriber) {
	var indexToRemove int
	for i, observer := range p.ObserverList {
		if observer == o {
			indexToRemove = i
			break
		}
	}
	p.ObserverList = append(
		p.ObserverList[:indexToRemove],
		p.ObserverList[indexToRemove+1:]...,
	)
}

func (p *Publisher) NotifyObservers(message string) {
	for _, observer := range p.ObserverList {
		observer.Notify(message)
	}
}

type TestObserver struct {
	ID      int
	Message string
}

func (p *TestObserver) Notify(m string) {
	fmt.Printf("Obderver %d: message '%s' received \n", p.ID, m)
	p.Message = m
}
