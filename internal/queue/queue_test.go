package queue

import (
	"fmt"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

const (
	url   = "nats://hyp:4222"
	topic = "MyTopic"
	msg   = "Hello there!!!"
)

func TestQueue_mock(t *testing.T) {

	queue := MockQueue{buff: Dict{}}

	_, err := queue.Connect(url)
	assert.NoError(t, err)

	err = queue.Publish(topic, []byte(msg))
	assert.NoError(t, err)

	_, err = queue.Subscribe(topic, func(str string) {
		fmt.Printf("*** Read value: %s \n", str)
		assert.Equal(t, msg, str)
	})
	assert.NoError(t, err)
}

func TestPub_nats(t *testing.T) {

	// Connect to a server
	nc, _ := nats.Connect(url)

	// Simple Publisher
	nc.Publish("foo", []byte("Hello World"))

	// Simple Async Subscriber
	nc.Subscribe("foo", func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})

	// Responding to a request message
	nc.Subscribe("request", func(m *nats.Msg) {
		m.Respond([]byte("answer is 42"))
	})

}
