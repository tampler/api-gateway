package cloudcontrol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/neurodyne-web-services/nws-sdk-go/services/cloudcontrol/api"
)

const (
	NATS_URL = "nats://hyp:4222"
	Topic    = "Command.Ingress"
)

type Command struct {
	service  string
	resource string
	action   string
}

type APIServer struct{}

// FIXME - reimplement this with a Pool
func MakeAPIServer() *APIServer {
	var srv APIServer

	return &srv
}

func (c *APIServer) GetMetrics(ctx echo.Context) error {
	return sendCloudControlError(ctx, http.StatusInternalServerError, "NYI - not yet implemented")
}

func (c *APIServer) PostV1(ctx echo.Context) error {
	// Extract API request from REST
	var req api.Request

	err := ctx.Bind(&req)
	if err != nil {
		return sendCloudControlError(ctx, http.StatusBadRequest, "Failed to parse input API request")
	}

	if err = ctx.Validate(req.Mandatory.Command); err != nil {
		return sendCloudControlError(ctx, http.StatusInternalServerError, err.Error())
	}

	var params []string

	if req.Options.Params != nil {
		params = *req.Options.Params
	}

	// Parse input command
	cmd := strings.Split(req.Mandatory.Command, delimiter)
	serviceName := cmd[1]
	resourceName := cmd[2]
	action := string(req.Mandatory.Action)

	fmt.Printf("**** Inputs: service - %s, resource - %s, action - %s \n", serviceName, resourceName, action)
	fmt.Println(params)

	comm := Command{
		service:  serviceName,
		resource: resourceName,
		action:   action,
	}

	res, err := sendRequestWithReply(comm)
	if err != nil {
		return sendCloudControlError(ctx, http.StatusInternalServerError, fmt.Sprintf("API error: %v", err))
	}

	fmt.Printf("*** API response: %s", string(res))

	return sendCloudControlError(ctx, http.StatusInternalServerError, "Failed to Execute command")

}

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.``
func sendCloudControlError(ctx echo.Context, code int, message string) error {
	petErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, petErr)
	return err
}

// sendRequestWithReply - sends a Cloud Control API command to subscribed executors
func sendRequestWithReply(cmd Command) ([]byte, error) {

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	enc.Encode(cmd)

	nc, err := nats.Connect(NATS_URL)
	if err != nil {
		return nil, err
	}
	defer nc.Close()

	reply, err := nc.Request("foo", []byte("I need help"), 4*time.Second)
	if err != nil {
		return nil, err
	}

	return reply.Data, nil
}
