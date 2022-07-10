package protoserver

import (
	"context"
	"fmt"
	"time"

	aj "github.com/choria-io/asyncjobs"
	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/token"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// protoServer is used to implement the Cloud Control GRPC server
type protoServer struct {
	cc.UnimplementedCloudControlServiceServer
	worker.APIServer
	pub *worker.Publisher
}

func MakeProtoServer(c *config.AppConfig, z *zap.SugaredLogger, ping, pong worker.QueueManager, pub *worker.Publisher) *protoServer {
	api := worker.MakeAPIServer(c, z, ping, pong)
	return &protoServer{
		cc.UnimplementedCloudControlServiceServer{}, api, pub,
	}
}

func (s *protoServer) UnaryCall(ctx context.Context, req *cc.APIRequest) (*cc.APIResponse, error) {
	var resp cc.APIResponse

	// Request ID validator
	requestID, err := uuid.Parse(req.JobID)
	if err != nil {
		s.Zl.Error(err)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// Command validator
	// This needs to match a REST validator string, thus build it from GRPC req
	cmdString := "NWS::" + req.Cmd.Service + "::" + req.Cmd.Resource

	if err := token.CommandValidator(cmdString); err != nil {
		s.Zl.Debugf(fmt.Sprintf("Failed to validate command: %v", err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// channel for observer syncronization
	done := make(chan bool)
	defer close(done)

	// Add observer to the server context
	observ := worker.MakeBusObserver(done)
	s.pub.AddObserver(requestID, &observ)
	defer s.pub.RemoveObserver(requestID)

	runtime := time.Duration(s.Cfg.Ajc.Timeout) * time.Minute

	task, err := aj.NewTask(s.Cfg.Ajc.Ingress.Topic, req, aj.TaskDeadline(time.Now().Add(runtime)))
	if err != nil {
		s.Zl.Debugf(fmt.Sprintf("Failed to create a task: %v", err))
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	s.Zl.Debugf("PING: push task %v", req.Cmd)

	// Submit a task into the PING queue
	err = s.Ping.Client.EnqueueTask(context.Background(), task)
	if err != nil {
		s.Zl.Debugf(fmt.Sprintf("Failed to submit a PING task: %v", err))
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	select {

	case <-time.After(time.Duration(s.Cfg.Sdk.JobTime) * time.Second):
		s.Zl.Errorf("FAIL: request timed out %v", req)

	case <-done:
		if len(observ.Err) > 0 {
			s.Zl.Errorf("Fail: error: %v", string(observ.Err))
		} else {
			s.Zl.Debugf("Success: response: %v", string(observ.Data))
		}
	}

	if observ.Err != "" {
		return nil, status.Errorf(codes.Aborted, observ.Err)
	}

	if observ.Data == nil {
		return nil, status.Errorf(codes.Aborted, "Empty Buffer")
	}

	return &resp, nil
}
