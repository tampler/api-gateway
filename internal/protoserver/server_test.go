package protoserver

import (
	"context"
	"fmt"
	"log"
	"testing"

	aj "github.com/choria-io/asyncjobs"

	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"github.com/neurodyne-web-services/nws-sdk-go/services/natstool"
	"github.com/stretchr/testify/assert"
)

const (
	CONFIG_PATH = "../../configs"
	CONFIG_NAME = "app"
)

func Test_unary(t *testing.T) {
	// Build a global config
	var cfg config.AppConfig

	if err := cfg.AppInit(CONFIG_NAME, CONFIG_PATH); err != nil {
		log.Fatalf("Config failed %s", err.Error())
	}

	// App logger
	logger, _ := logging.MakeLogger(cfg.Log.Verbosity, cfg.Log.Output)
	defer logger.Sync()
	zl := logger.Sugar()

	// Connect to NATS
	nc, err := natstool.MakeNatsConnect()
	if err != nil {
		zl.Fatalf("NATS connect failed %s \n", err.Error())
	}

	// Input queue
	pingClient, err := aj.NewClient(
		aj.NatsConn(nc),
		aj.BindWorkQueue(cfg.Ajc.Ingress.Name),
		aj.ClientConcurrency(cfg.Ajc.Ingress.Concurrency),
		aj.PrometheusListenPort(cfg.Ajc.Ingress.MetricsPort),
		aj.RetryBackoffPolicy(aj.RetryLinearOneMinute))

	if err != nil {
		zl.Fatal(err)
	}

	// Output queue
	pongClient, err := aj.NewClient(
		aj.NatsConn(nc),
		aj.BindWorkQueue(cfg.Ajc.Egress.Name),
		aj.ClientConcurrency(cfg.Ajc.Egress.Concurrency),
		aj.PrometheusListenPort(cfg.Ajc.Egress.MetricsPort),
		aj.RetryBackoffPolicy(aj.RetryLinearOneMinute))

	if err != nil {
		zl.Fatal(err)
	}

	// Create queue routers
	pingRouter := aj.NewTaskRouter()
	pongRouter := aj.NewTaskRouter()

	pingMgr := worker.MakeQueueManager(pingClient, pingRouter)
	pongMgr := worker.MakeQueueManager(pongClient, pongRouter)

	pub := worker.MakePublisher(pongMgr, zl, map[uuid.UUID]worker.Subscriber{})
	pub.AddHandlers(cfg.Ajc.Egress.Topic)

	srv := MakeProtoServer(&cfg, zl, pingMgr, pongMgr, &pub)

	cmd := cc.APICommand{
		Service:  "EC2",
		Resource: "SSHKeypair",
		Action:   "List",
		Params:   []string{},
	}

	req := cc.APIRequest{
		JobID: uuid.NewString(),
		Cmd:   &cmd,
	}

	res, err := srv.UnaryCall(context.Background(), &req)
	assert.NoError(t, err)

	fmt.Println(res)
}
