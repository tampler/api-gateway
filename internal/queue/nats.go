package queue

import (
	"github.com/nats-io/nats.go"
)

type NatsQueue struct {
	client *nats.Conn
}

func (nq NatsQueue) Connect(topic string) error {

	nc, err := nats.Connect(NATS_URL)
	if err != nil {
		return err
	}
	defer nc.Close()

	nq.client = nc

	return nil
}

func (nq NatsQueue) Publish(topic string, data []byte) error {
	return nq.client.Publish(topic, data)
}

func (nq NatsQueue) Subscribe(topic string, cb nats.MsgHandler) error {

	_, err := nq.client.Subscribe(topic, cb)
	if err != nil {
		return err
	}

	return nil
}
