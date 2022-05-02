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

	cc := ctx.(*MyContext)
	cc.Foo()

	// Add observer
	observ := TestObserver{222, nil}

	cc.pub.AddSubscriber(&observ)

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
	cmd := APIRequestMessage{
		Cmd: APIRequestCommand{
			JobID:    uuid.NewV4().String(),
			Service:  serviceName,
			Resource: resourceName,
			Action:   action,
			Params:   params,
		},
	}

	fmt.Printf("*** API Server called %v \n", cmd)

	task, err := aj.NewTask(topic, cmd.Cmd, aj.TaskDeadline(time.Now().Add(time.Hour)))
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to create a task: %v \n", err))
	}

	// Submit a task into the PING queue
	err = s.ping.client.EnqueueTask(context.Background(), task)
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to submit a PING task"))
	}

	// Block after creating a PING event
	time.Sleep(3 * time.Second)

	return nil
}

func sendResponse(cfg handlerConfig) error {

	// Repack to the full Runner Result
	out := APIResponseMessage{
		Service: cfg.service,
		Api:     cfg.resource,
		Data:    cfg.data,
	}

	buf, err := json.Marshal(&out)
	if err != nil {
		return sendAPIError(cfg.ctx, http.StatusInternalServerError, "Failed to serialize Runner Response")
	}

	fmt.Printf("*** Sending buf: %v\n", string(buf))

	// Now, we have to return the Runner response
	err = cfg.ctx.JSONBlob(http.StatusCreated, buf)
	if err != nil {
		return sendAPIError(cfg.ctx, http.StatusInternalServerError, "Failed to send response")
	}

	return nil
}
