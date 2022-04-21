package apiserver

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
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
func MakeAPIServer(nc *nats.Conn, c *config.AppConfig, z *zap.Logger) *APIServer {
	srv := APIServer{
		nats: nc,
		zl:   z,
		cfg:  c,
	}

	return &srv
}

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
	comm := APIRequestMessage{
		Cmd: APIRequestCommand{
			Service:  serviceName,
			Resource: resourceName,
			Action:   action,
			Params:   params,
		},
		Cfg: NatsConfig{
			Timeout: s.cfg.Nats.Timeout,
			Server:  s.cfg.Nats.Server,
			Topic:   s.cfg.Nats.Topic,
		},
	}

	res, err := s.sendRequestWithReply(comm)
	if err != nil {
		return sendAPIError(ctx, http.StatusInternalServerError, fmt.Sprintf("API error: %v", err))
	}

	// Repack to the full Runner Result
	out := APIResponseMessage{
		Service: serviceName,
		Api:     resourceName,
		Data:    res,
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

// sendRequestWithReply - sends a Cloud Control API command to subscribed executors
func (s *APIServer) sendRequestWithReply(msg APIRequestMessage) ([]byte, error) {

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	enc.Encode(msg)

	reply, err := s.nats.Request(msg.Cfg.Topic, buff.Bytes(), time.Duration(msg.Cfg.Timeout)*time.Second)
	if err != nil {
		return nil, err
	}

	return reply.Data, nil
}
