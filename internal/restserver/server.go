package restserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	aj "github.com/choria-io/asyncjobs"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"github.com/neurodyne-web-services/api-gateway/pkg/rest"
)

// protoServer is used to implement the Cloud Control REST server
type restServer struct {
	worker.APIServer
}

// MakeRestServer - APIServer factory
func MakeRestServer(c *config.AppConfig, z *zap.SugaredLogger, ping, pong worker.QueueManager) *restServer {
	api := worker.MakeAPIServer(c, z, ping, pong)
	return &restServer{api}
}

func (s *restServer) GetMetrics(ctx echo.Context) error {
	return sendAPIError(ctx, http.StatusInternalServerError, "NYI - not yet implemented")
}

func (s *restServer) PostV1(ctx echo.Context) error {

	// Apply custom context
	cc := ctx.(*MyContext)

	// Create and store request ID
	requestID := uuid.New()

	done := make(chan bool)
	defer close(done)

	// Add observer
	observ := worker.MakeBusObserver(done)
	cc.pub.AddObserver(requestID, &observ)
	defer cc.pub.RemoveObserver(requestID)

	// Extract API request from REST
	var req rest.Req

	err := ctx.Bind(&req)
	if err != nil {
		return sendAPIError(ctx, http.StatusBadRequest, "Failed to parse input API request")
	}

	// Validate request fields
	if err = ctx.Validate(req.Mandatory.Command); err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, err.Error())
	}

	params := req.Mandatory.Params

	// Parse input command
	msg := strings.Split(req.Mandatory.Command, delimiter)
	serviceName := msg[1]
	resourceName := msg[2]

	// Save User ID from JWT for the Session Login event
	if serviceName == "Session" && resourceName == "Login" {
		params = append(params, cc.info.ID)
	}

	// Extract API command
	cmd := APIRequest{
		JobID: requestID.String(),
		Cmd: APICommand{
			Service:  serviceName,
			Resource: resourceName,
			Action:   string(req.Mandatory.Action),
			Params:   params,
		},
	}

	runtime := time.Duration(cc.cfg.Ajc.Timeout) * time.Minute

	task, err := aj.NewTask(cc.cfg.Ajc.Ingress.Topic, cmd, aj.TaskDeadline(time.Now().Add(runtime)))
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to create a task: %v", err))
	}

	cc.zl.Debugf("PING adding task %v", cmd)

	// Submit a task into the PING queue
	err = s.Ping.Client.EnqueueTask(context.Background(), task)
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to submit a PING task: %v", err))
	}

	select {

	case <-time.After(time.Duration(cc.cfg.Sdk.JobTime) * time.Second):
		cc.zl.Errorf("FAIL: request timed out %v", req)

	case <-done:
		if len(observ.Err) > 0 {
			cc.zl.Errorf("Fail: error: %v", string(observ.Err))
		} else {
			cc.zl.Debugf("Success: response: %v", string(observ.Data))
		}
	}

	if observ.Err != "" {
		return sendAPIError(ctx, http.StatusInternalServerError, observ.Err)
	}

	if observ.Data == nil {
		return sendAPIError(ctx, http.StatusInternalServerError, "Empty buffer")
	}

	return sendResponse(cc, observ.Data, serviceName, resourceName)
}

func sendResponse(ctx *MyContext, data []byte, service, resource string) error {

	// Repack to the full Runner Result
	out := APIResponse{
		JobID: uuid.NewString(),
		Err:   "",
		Data:  data,
	}

	buf, err := json.Marshal(&out)
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, "Failed to serialize Runner Response")
	}

	// Now, we have to return the Runner response
	err = ctx.JSONBlob(http.StatusCreated, buf)
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, "Failed to send response")
	}

	return nil
}
