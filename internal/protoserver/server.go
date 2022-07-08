package protoserver

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
)

// grpcServer is used to implement the Cloud Control GRPC server
type Server struct {
	cc.UnimplementedCloudControlServiceServer
}

func (s *Server) UnaryCall(context.Context, *cc.APIRequest) (*cc.APIResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "TBD: impl the service")
}
