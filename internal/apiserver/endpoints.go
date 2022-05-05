package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	aj "github.com/choria-io/asyncjobs"

	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/nws-sdk-go/services/cloudcontrol/api"
	uuid "github.com/satori/go.uuid"
)

const (
	topic = "sdk::ec2"
)

func (s *APIServer) GetMetrics(ctx echo.Context) error {
	return sendAPIError(ctx, http.StatusInternalServerError, "NYI - not yet implemented")
}

func (s *APIServer) PostV1(ctx echo.Context) error {

	// Apply custom context
	cc := ctx.(*MyContext)

	// Create and store request ID
	requestID := uuid.NewV4()

	done := make(chan bool)
	defer close(done)

	// Add observer
	observ := MakeBusObserver(requestID, cc.zl, done)
	cc.pub.AddObserver(requestID, &observ)
	defer cc.pub.RemoveObserver(requestID)

	// Extract API request from REST
	var req api.Request

	err := ctx.Bind(&req)
	if err != nil {
		return sendAPIError(ctx, http.StatusBadRequest, "Failed to parse input API request")
	}

	// Validate request fields
	if err = ctx.Validate(req.Mandatory.Command); err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, err.Error())
	}

	var params []string

	if req.Options.Params != nil {
		params = *req.Options.Params
	}

	// Parse input command
	msg := strings.Split(req.Mandatory.Command, delimiter)
	serviceName := msg[1]
	resourceName := msg[2]
	action := string(req.Mandatory.Action)

	// Extract API command
	cmd := APIRequest{
		JobID: requestID,
		Cmd: APICommand{
			Service:  serviceName,
			Resource: resourceName,
			Action:   action,
			Params:   params,
		},
	}

	task, err := aj.NewTask(topic, cmd, aj.TaskDeadline(time.Now().Add(time.Hour)))
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to create a task: %v \n", err))
	}

	cc.zl.Debugf("PING adding task %v \n", cmd)

	// Submit a task into the PING queue
	err = s.ping.client.EnqueueTask(context.Background(), task)
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to submit a PING task"))
	}

	select {

	case <-time.After(8 * time.Second):
		cc.zl.Errorf("FAIL: to execute command")

	case <-done:
		cc.zl.Debugf("Message: %v\n", string(observ.data))

	}

	if observ.err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, err.Error())
	}

	if observ.data == nil {
		return sendAPIError(ctx, http.StatusInternalServerError, "Empty buffer")
	}

	return sendResponse(cc, observ.data, serviceName, resourceName)
}

func sendResponse(ctx *MyContext, data []byte, service, resource string) error {

	// Repack to the full Runner Result
	out := APIResponse{
		JobID: uuid.Nil,
		Err:   nil,
		Data:  data,
	}

	buf, err := json.Marshal(&out)
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, "Failed to serialize Runner Response")
	}

	ctx.zl.Debugf("Sending buf: %v\n", string(buf))

	// Now, we have to return the Runner response
	err = ctx.JSONBlob(http.StatusCreated, buf)
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, "Failed to send response")
	}

	return nil
}
