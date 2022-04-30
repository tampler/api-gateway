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

	task, err := aj.NewTask(topic, cmd, aj.TaskDeadline(time.Now().Add(time.Hour)))
	if err != nil {
		return err
	}

	err = s.ping.client.EnqueueTask(context.Background(), task)
	if err != nil {
		return err
	}

	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("API error: %v", err))
	}

	// Repack to the full Runner Result
	out := APIResponseMessage{
		Service: serviceName,
		Api:     resourceName,
		Data:    nil,
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

	// Return no error. This refers to the handler. Even if we return an HTTP
	// error, but everything else is working properly, tell Echo that we serviced
	// the error. We should only return errors from Echo handlers if the actual
	// servicing of the error on the infrastructure level failed. Returning an
	// HTTP/400 or HTTP/500 from here means Echo/HTTP are still working, so
	// return nil.
	return nil
}
