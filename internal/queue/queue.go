package queue

const (
	NATS_URL = "nats://hyp:4222"
)

// Dict - dictionary
type Dict map[string][]byte

// SubCallback - publish method callback
type SubCallback[A any] func(A)

// Queue - generic queue impl to support Mock, Nats, Kafka etc
type Queue[A any] interface {
	Connect(string, ...string) error
	Publish(string, []byte) error
	Subscribe(string, func(*A)) error
}
