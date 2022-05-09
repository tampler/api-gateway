package nats

import (
	"os"

	"github.com/nats-io/nats.go"
)

// MakeNatsConnect - establishes a connection to NATS JetStream server
func MakeNatsConnect() (*nats.Conn, error) {

	natsUrl := os.Getenv("NATS_URL")
	natsUser := os.Getenv("NATS_USER")
	natsPass := os.Getenv("NATS_PASS")
	natsConnString := natsUser + ":" + natsPass + "@" + natsUrl

	return nats.Connect(natsConnString, nats.UseOldRequestStyle())
}
