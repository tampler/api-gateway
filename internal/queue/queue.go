package queue

const (
	NATS_URL = "nats://hyp:4222"
)

type Dict map[string][]byte
type SubCallback[A any] func(A)

// type Publisher[A, B any] interface {
// 	Connect(string, ...A) (B, error)
// 	Publish(string, []byte) error
// }
type Queue interface {
	Connect(string, ...string) (string, error)
	Publish(string, []byte) error
	Subscribe(string, SubCallback[string]) (string, error)
}
