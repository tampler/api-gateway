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
func MakeAPIServer(nc *nats.Conn, c *config.AppConfig, z *zap.Logger) *APIServer {
	srv := APIServer{
		nats: nc,
		zl:   z,
		cfg:  c,
	}

	return &srv
}

func (s *APIServer) GetMetrics(ctx echo.Context) error {
	return sendCloudControlError(ctx, http.StatusInternalServerError, "NYI - not yet implemented")
}

func (s *APIServer) PostV1(ctx echo.Context) error {

	// Extract API request from REST
	var req api.Request

	err := ctx.Bind(&req)
	if err != nil {
		return sendCloudControlError(ctx, http.StatusBadRequest, "Failed to parse input API request")
	}

	// Validate request fields
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

	// Extract API command
	comm := APIMessage{
		Cmd: APICommand{
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
		return sendCloudControlError(ctx, http.StatusInternalServerError, fmt.Sprintf("API error: %v", err))
	}

	fmt.Printf("*** Exec response: %s \n", string(res))

	return sendCloudControlError(ctx, http.StatusInternalServerError, "Failed to Execute command")
}

// sendRequestWithReply - sends a Cloud Control API command to subscribed executors
func (s *APIServer) sendRequestWithReply(msg APIMessage) ([]byte, error) {

	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	enc.Encode(msg)

	reply, err := s.nats.Request(msg.Cfg.Topic, buff.Bytes(), time.Duration(msg.Cfg.Timeout)*time.Second)
	if err != nil {
		return nil, err
	}

	return reply.Data, nil
}
