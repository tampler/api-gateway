package protoserver

import (
	"context"

	"github.com/neurodyne-web-services/api-gateway/internal/token"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcServer is used to implement the Cloud Control GRPC server
type Server struct {
	cc.UnimplementedCloudControlServiceServer
	worker.APIServer
}

func (s *Server) UnaryCall(ctx context.Context, req *cc.APIRequest) (*cc.APIResponse, error) {
	var resp cc.APIResponse

	// This needs to match a REST validator string, thus build it from GRPC req
	cmdString := "NWS::" + req.Cmd.Service + "::" + req.Cmd.Resource + "::" + req.Cmd.Action

	if err := token.CommandValidator(cmdString); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &resp, nil
}
