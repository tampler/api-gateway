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

func TestQueue_nats(t *testing.T) {

	queue := NatsQueue{}

	err := queue.Connect(NATS_URL)
	assert.NoError(t, err)
	assert.NotNil(t, queue.client)

	err = queue.Publish(topic, []byte(msg))
	assert.NoError(t, err)

	err = queue.Subscribe(topic, func(m *nats.Msg) {
		fmt.Printf("*** Read value: %v \n", m)
		assert.Equal(t, msg, m)
	})
	assert.NoError(t, err)
}
