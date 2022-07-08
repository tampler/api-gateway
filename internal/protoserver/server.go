package protoserver

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
)

// grpcServer is used to implement the Cloud Control GRPC server
type Server struct {
	cc.UnimplementedCloudControlServiceServer
	zl   *zap.SugaredLogger
	cfg  *config.AppConfig
	ping worker.QueueManager
	pong worker.QueueManager
}

func (s *Server) UnaryCall(context.Context, *cc.APIRequest) (*cc.APIResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "TBD: impl the service")
}
