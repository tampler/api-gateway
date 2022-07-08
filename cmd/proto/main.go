package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"

	aj "github.com/choria-io/asyncjobs"
	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/neurodyne-web-services/api-gateway/internal/protoserver"
	"github.com/neurodyne-web-services/api-gateway/internal/token"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"github.com/neurodyne-web-services/nws-sdk-go/services/natstool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	CONFIG_PATH = "./configs"
	CONFIG_NAME = "app"
)

func main() {

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

	// GRPC server
	var opts *[]grpc.ServerOption

	opts, err = buildServerOpts(&cfg)
	if err != nil {
		zl.Fatal(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Grpc.Port))
	if err != nil {
		zl.Fatal(err)
	}

	s := grpc.NewServer(*opts...)

	cc.RegisterCloudControlServiceServer(s, protoserver.MakeProtoServer(&cfg, zl, pingMgr, pongMgr, pub))

	showDebugInfo(zl.Desugar(), &cfg)
	if err := s.Serve(lis); err != nil {
		zl.Fatalf("failed to serve: %v", err)
	}
}

// buildServerOpts - returns a GRPC server options
func buildServerOpts(cfg *config.AppConfig) (*[]grpc.ServerOption, error) {

	if !cfg.Grpc.AuthEnabled {
		opts := []grpc.ServerOption{}
		return &opts, nil
	}

	cert, err := tls.LoadX509KeyPair(cfg.Grpc.CertFile, cfg.Grpc.KeyFile)
	if err != nil {
		return nil, err
	}

	opts := []grpc.ServerOption{
		// Intercept request to check the token.
		grpc.UnaryInterceptor(token.ValidateToken),
		// Enable TLS for all incoming connections.
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	}

	return &opts, nil
}

// showDebugInfo - this prints envs to ease deployment and debug
func showDebugInfo(zl *zap.Logger, cfg *config.AppConfig) {
	zl.Info("NATS URL: ", zap.String("NATS_URL", os.Getenv("NATS_URL")))
	zl.Info("GRPC URL: ", zap.Int("localhost", cfg.Grpc.Port))
	zl.Info("GRPC Secure: ", zap.Bool("auth", cfg.Grpc.AuthEnabled))
	zl.Info("Job timeout:", zap.Int("timeout, sec", cfg.Sdk.JobTime))
}
