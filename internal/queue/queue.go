package queue

const (
	NATS_URL = "nats://hyp:4222"
)

type Dict map[string][]byte
type SubCallback[A any] func(A)

type Queue[A any] interface {
	Connect(string, ...string) error
	Publish(string, []byte) error
	Subscribe(string, func(*A)) error
}
