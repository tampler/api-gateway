package protoserver

import (
	"fmt"

	aj "github.com/choria-io/asyncjobs"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CONFIG_PATH = "../../configs"
	CONFIG_NAME = "app"
	buck        = "ec2buck"
)

type storageServer struct {
	kv nats.KeyValue
}

func MakeStorageServer(nc *nats.Conn) (storageServer, error) {

	var serv storageServer

	// Setup a JetStream
	js, err := nc.JetStream()
	if err != nil {
		return serv, fmt.Errorf("NATS JetStream failed %s \n", err.Error())
	}

	kv, err := js.KeyValue(buck)
	if err != nil {
		return serv, fmt.Errorf("NATS KeyValue failed %s \n", err.Error())
	}

	serv.kv = kv

	return serv, nil
}

// buildProtoServer - generates a protobuf server with NATS support
func buildProtoServer(nc *nats.Conn, cfg config.AppConfig, zl *zap.SugaredLogger) (*protoServer, error) {

	// Input queue
	pingClient, err := aj.NewClient(
		aj.NatsConn(nc),
		aj.BindWorkQueue(cfg.Ajc.Ingress.Name),
		aj.ClientConcurrency(cfg.Ajc.Ingress.Concurrency),
		aj.PrometheusListenPort(cfg.Ajc.Ingress.MetricsPort),
		aj.RetryBackoffPolicy(aj.RetryLinearOneMinute))

	if err != nil {
		zl.Error(err)
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	// Output queue
	pongClient, err := aj.NewClient(
		aj.NatsConn(nc),
		aj.BindWorkQueue(cfg.Ajc.Egress.Name),
		aj.ClientConcurrency(cfg.Ajc.Egress.Concurrency),
		aj.PrometheusListenPort(cfg.Ajc.Egress.MetricsPort),
		aj.RetryBackoffPolicy(aj.RetryLinearOneMinute))

	if err != nil {
		zl.Error(err)
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	// Create queue routers
	pingRouter := aj.NewTaskRouter()
	pongRouter := aj.NewTaskRouter()

	pingMgr := worker.MakeQueueManager(pingClient, pingRouter)
	pongMgr := worker.MakeQueueManager(pongClient, pongRouter)

	pub := worker.MakePublisher(pongMgr, zl, map[uuid.UUID]worker.Subscriber{})
	pub.AddHandlers(cfg.Ajc.Egress.Topic)

	return MakeProtoServer(&cfg, zl, pingMgr, pongMgr, &pub), nil
}
