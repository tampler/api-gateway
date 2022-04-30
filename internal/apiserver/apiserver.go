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

type ContextMap = map[string]echo.Context

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

	// var cmap ContextMap

	// Store context for further user
	// cmap[cmd.Cmd.JobID] = ctx
	echoCTX := &ctx

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

	// Create a PONG handler
	err = s.pong.router.HandleFunc(topic, func(ctx context.Context, _ aj.Logger, t *aj.Task) (interface{}, error) {

		// bytes, err := decodeJSONBytes(t.Payload)
		// encData, err := jsonparser.GetString(t.Payload)
		// if err != nil {
		// 	return nil, err
		// }

		// str, _ := base64.StdEncoding.DecodeString(string(encData))

		// fmt.Printf("*** PONG API Response: %v\n", string(str))

		err = sendResponse(echoCTX, t.Payload, serviceName, resourceName)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch a PONG task"))
	}

	ch := make(chan error, 2)

	go s.ping.Run(context.Background(), ch)
	go s.pong.Run(context.Background(), ch)

	pingErr, pongErr := <-ch, <-ch

	if pingErr != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to run a PING manager: %v", pingErr.Error()))
	}

	if pongErr != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to run a PONG manager: %v", pongErr.Error()))
	}

	return nil
}

func sendResponse(ctx *echo.Context, bytes []byte, service, resource string) error {

	// Repack to the full Runner Result
	out := APIResponseMessage{
		Service: service,
		Api:     resource,
		Data:    bytes,
	}

	buf, err := json.Marshal(&out)
	if err != nil {
		return sendAPIError(*ctx, http.StatusInternalServerError, "Failed to serialize Runner Response")
	}

	fmt.Printf("*** Sending JSON BLOB for context %v \n", *ctx)

	// Now, we have to return the Runner response
	err = (*ctx).JSONBlob(http.StatusCreated, buf)
	if err != nil {
		return sendAPIError(*ctx, http.StatusInternalServerError, "Failed to send response")
	}

	fmt.Printf("*** Sending JSON BLOB Done \n")

	// Return no error. This refers to the handler. Even if we return an HTTP
	// error, but everything else is working properly, tell Echo that we serviced
	// the error. We should only return errors from Echo handlers if the actual
	// servicing of the error on the infrastructure level failed. Returning an
	// HTTP/400 or HTTP/500 from here means Echo/HTTP are still working, so
	// return nil.
	return nil
}
