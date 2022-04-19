package queue

import "fmt"

type MockQueue struct {
	buff Dict
}

func (mp MockQueue) Connect(url string, opts ...string) (string, error) {
	fmt.Printf("Mock Pub: Connected to URL %v \n", url)
	return "", nil
}

func (mp MockQueue) Publish(topic string, data []byte) error {
	mp.buff[topic] = data
	fmt.Printf("Mock Pub: Publish success... \n")
	return nil
}

func (mp MockQueue) Subscribe(topic string, cb SubCallback[string]) (string, error) {
	msg := mp.buff[topic]
	cb(string(msg))
	return "", nil
}
