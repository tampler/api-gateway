package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"

	aj "github.com/choria-io/asyncjobs"
	"github.com/neurodyne-web-services/api-gateway/internal/apiserver"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/neurodyne-web-services/api-gateway/internal/token"
	"github.com/neurodyne-web-services/api-gateway/pkg/greeter"
	"github.com/neurodyne-web-services/nws-sdk-go/services/natstool"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	CONFIG_PATH = "./configs"
	CONFIG_NAME = "app"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	greeter.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *greeter.HelloRequest) (*greeter.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &greeter.HelloReply{Message: "Hello " + in.GetName()}, nil
}

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

	pingMgr := apiserver.MakeQueueManager(pingClient, pingRouter)
	pongMgr := apiserver.MakeQueueManager(pongClient, pongRouter)

	// Create an instance of our handler which satisfies the generated interface
	_ = apiserver.MakeAPIServer(&cfg, zl, pingMgr, pongMgr)

	pub := apiserver.MakePublisher(pongMgr, zl, map[uuid.UUID]apiserver.Subscriber{})
	pub.AddHandlers(cfg.Ajc.Egress.Topic)

	showDebugInfo(zl.Desugar(), &cfg)

	// GRPC server
	var opts *[]grpc.ServerOption

	if cfg.Http.AuthEnabled {
		opts, err = buildSecureOpts(&cfg)
		if err != nil {
			zl.Fatal(err)
		}
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Grpc.Port))
	if err != nil {
		zl.Fatal(err)
	}

	s := grpc.NewServer(*opts...)

	greeter.RegisterGreeterServer(s, &server{})

	zl.Infof("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		zl.Fatalf("failed to serve: %v", err)
	}
}

// showDebugInfo - this prints envs to ease deployment and debug
func showDebugInfo(zl *zap.Logger, cfg *config.AppConfig) {
	zl.Info("NATS URL: ", zap.String("NATS_URL", os.Getenv("NATS_URL")))
	zl.Info("Job timeout:", zap.Int("timeout, sec", cfg.Sdk.JobTime))
}

func buildSecureOpts(cfg *config.AppConfig) (*[]grpc.ServerOption, error) {

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
