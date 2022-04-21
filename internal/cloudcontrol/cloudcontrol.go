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
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"github.com/neurodyne-web-services/nws-sdk-go/services/cloudcontrol/api"
	"go.uber.org/zap"
)

// MakeAPIServer - APIServer factory
func MakeAPIServer(c *config.AppConfig, z *zap.Logger) *APIServer {
	srv := APIServer{
		zl:  z,
		cfg: c,
	}

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
	msg := strings.Split(req.Mandatory.Command, delimiter)
	serviceName := msg[1]
	resourceName := msg[2]
	action := string(req.Mandatory.Action)

	fmt.Printf("**** API GW --- Inputs: service - %s, resource - %s, action - %s \n", serviceName, resourceName, action)
	fmt.Println(params)

	comm := APIMessage{
		Service:  serviceName,
		Resource: resourceName,
		Action:   action,
		Cfg: NatsConfig{
			Timeout: c.cfg.Nats.Timeout,
			Server:  c.cfg.Nats.Server,
			Topic:   c.cfg.Nats.Topic,
		},
	}

	res, err := sendRequestWithReply(comm)
	if err != nil {
		return sendCloudControlError(ctx, http.StatusInternalServerError, fmt.Sprintf("API error: %v", err))
	}

	fmt.Printf("*** Exec response: %s \n", string(res))

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
func sendRequestWithReply(msg APIMessage) ([]byte, error) {

	// fmt.Printf("****NATS server %s topic %s \n", msg.Cfg.server, msg.Cfg.topic)
	fmt.Printf(">>> Command: %v \n", msg)

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	enc.Encode(msg)

	nc, err := nats.Connect(msg.Cfg.Server)
	if err != nil {
		return nil, err
	}
	defer nc.Close()

	reply, err := nc.Request(msg.Cfg.Topic, buff.Bytes(), time.Duration(msg.Cfg.Timeout)*time.Second)
	if err != nil {
		return nil, err
	}

	nc.Drain()

	return reply.Data, nil
}
