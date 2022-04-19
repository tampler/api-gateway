package queue

import "fmt"

type MockQueue struct {
	buff Dict
}

func (mq MockQueue) Connect(url string, opts ...string) (string, error) {
	fmt.Printf("Mock Queue: Connected to URL %v \n", url)
	return "", nil
}

func (mq MockQueue) Publish(topic string, data []byte) error {
	mq.buff[topic] = data
	fmt.Printf("Mock Queue: Publish success... \n")
	return nil
}

func (mq MockQueue) Subscribe(topic string, cb SubCallback[string]) (string, error) {
	msg := mq.buff[topic]
	cb(string(msg))
	return "", nil
}
