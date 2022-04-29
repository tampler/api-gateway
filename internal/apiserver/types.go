package apiserver

import (
	aj "github.com/choria-io/asyncjobs"
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"go.uber.org/zap"
)

// APIServer - top level execution engine
type APIServer struct {
	zl     *zap.Logger
	cfg    *config.AppConfig
	ping   *aj.Client
	pong   *aj.Client
	router *aj.Mux
}

// APICommand - API command
type APIRequestCommand struct {
	Service  string
	Resource string
	Action   string
	Params   []string
}

// APIMessage - message to be processed by SDK
type APIRequestMessage struct {
	Cmd APIRequestCommand
}

// APIResponseMessage - final JSON output from SDK
type APIResponseMessage struct {
	Service string `json:"service"`
	Api     string `json:"api"`
	Data    []byte `json:"data"`
}

type AsyncProcessor interface {
	Process() (interface{}, error)
}
