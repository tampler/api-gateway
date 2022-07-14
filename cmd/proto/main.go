package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/neurodyne-web-services/api-gateway/internal/protoserver"
	"github.com/neurodyne-web-services/api-gateway/internal/token"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"github.com/neurodyne-web-services/nws-sdk-go/services/natstool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	CONFIG_PATH = "./configs"
	CONFIG_NAME = "app"
)

func main() {
	ctx := context.Background()

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

	// Proto server with Nats infra
	server, err := protoserver.BuildProtoServer(ctx, nc, cfg, zl)
	if err != nil {
		zl.Fatal(err)
	}

	// GRPC server
	var opts *[]grpc.ServerOption

	opts, err = buildServerOpts(&cfg)
	if err != nil {
		zl.Fatal(err)
	}

	// Register GRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Grpc.Port))
	if err != nil {
		zl.Fatal(err)
	}

	s := grpc.NewServer(*opts...)
	cc.RegisterCloudControlServiceServer(s, server)

	// Register reflection service on gRPC server if Reflection enabled
	// Do NOT use in PROD !!!
	if cfg.Grpc.ReflectEnabled {
		reflection.Register(s)
		if err := s.Serve(lis); err != nil {
			zl.Fatal(err)
		}
	}

	showDebugInfo(zl.Desugar(), &cfg)
	if err := s.Serve(lis); err != nil {
		zl.Fatalf("failed to serve: %v", err)
	}
}

// buildServerOpts - returns a GRPC server options
func buildServerOpts(cfg *config.AppConfig) (*[]grpc.ServerOption, error) {

	if !cfg.Grpc.TLSEnabled {
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
	zl.Info("GRPC URL: ", zap.Int("localhost", cfg.Grpc.Port))
	zl.Info("NATS URL: ", zap.String("NATS_URL", os.Getenv("NATS_URL")))
	zl.Info("GRPC Auth: ", zap.Bool("TLS", cfg.Grpc.TLSEnabled))
	zl.Info("GRPC Reflect: ", zap.Bool("Reflection", cfg.Grpc.ReflectEnabled))
	zl.Info("Job timeout:", zap.Int("timeout, sec", cfg.Sdk.JobTime))
}
