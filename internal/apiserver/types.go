package apiserver

import (
	aj "github.com/choria-io/asyncjobs"
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"go.uber.org/zap"
)

// APIServer - top level execution engine
type APIServer struct {
	zl   *zap.Logger
	cfg  *config.AppConfig
	inc  *aj.Client
	outc *aj.Client
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
