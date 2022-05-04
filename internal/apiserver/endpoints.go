package apiserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	aj "github.com/choria-io/asyncjobs"

	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
	"github.com/neurodyne-web-services/nws-sdk-go/services/cloudcontrol/api"
	uuid "github.com/satori/go.uuid"
)

const (
	topic = "sdk::ec2"
)

var currentID uuid.UUID

func (s *APIServer) GetMetrics(ctx echo.Context) error {
	return sendAPIError(ctx, http.StatusInternalServerError, "NYI - not yet implemented")
}

func (s *APIServer) PostV1(ctx echo.Context) error {

	// Apply custom context
	cc := ctx.(*MyContext)

	// Create and store request ID
	requestID := uuid.NewV4()
	currentID = requestID

	done := make(chan bool)

	// Add observer
	observ := MakeBusObserver(requestID, nil, cc.zl, done)
	cc.pub.AddObserver(requestID, &observ)
	defer cc.pub.RemoveObserver(currentID)

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
		JobID: requestID.String(),
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

	cc.zl.Debug("PING task added")

	select {
	case <-time.After(5 * time.Second):
		cc.zl.Errorf("FAIL: to execute command")
	case <-done:
		cc.zl.Debugf("Message: %v\n", string(observ.data))
	}

	return sendResponse(cc, observ.data, serviceName, resourceName)
}

func sendResponse(ctx *MyContext, data []byte, service, resource string) error {

	// Repack to the full Runner Result
	out := APIResponseMessage{
		Service: service,
		Api:     resource,
		Data:    data,
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

func decodeJSONBytes(bytes []byte) ([]byte, error) {

	// Base-64 encoded string after marshalling []byte
	encData, err := jsonparser.GetString(bytes)
	if err != nil {
		return nil, fail.Error500(err.Error())
	}

	return base64.StdEncoding.DecodeString(string(encData))
}
