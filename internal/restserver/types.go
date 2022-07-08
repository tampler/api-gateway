package restserver

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"go.uber.org/zap"
)

const (
	// dummy user ID for testing
	userID = "12eb8d3e-ea8a-4aa1-9226-5d3762aa668e"

	// Cloud Control Commands
	sessCommand      = "NWS::Session::Login"
	zoneCommand      = "NWS::EC2::Zone"
	domCommand       = "NWS::EC2::Domain"
	accCommand       = "NWS::EC2::Account"
	sshCommand       = "NWS::EC2::SSHKeypair"
	vpcCommand       = "NWS::EC2::VPC"
	vpcOfferCommand  = "NWS::EC2::VPCOffer"
	netCommand       = "NWS::EC2::Network"
	netOfferCommand  = "NWS::EC2::NetOffer"
	osOfferCommand   = "NWS::EC2::OSOffer"
	tmplCommand      = "NWS::EC2::Template"
	instCommand      = "NWS::EC2::Instance"
	instOfferCommand = "NWS::EC2::InstOffer"
	aclCommand       = "NWS::EC2::ACL"
	aclrCommand      = "NWS::EC2::ACLRule"
)

// SubMap - event subscriber map
type SubMap = map[uuid.UUID]Subscriber

// user info from JWT
type UserInfo struct {
	ID string
}

// MyContext - custom echo context
type MyContext struct {
	echo.Context
	cfg  config.AppConfig
	pub  *Publisher
	zl   *zap.SugaredLogger
	info UserInfo
}

// MakeMyContext - factory to create a context
func MakeMyContext(c echo.Context, cfg config.AppConfig, pub *Publisher, zl *zap.SugaredLogger, info UserInfo) *MyContext {
	return &MyContext{
		c, cfg, pub, zl, info,
	}
}

type testServer struct {
	echo *echo.Echo
	kv   nats.KeyValue
}

// APICommand - cloud control command
type APICommand struct {
	Service  string
	Resource string
	Action   string
	Params   []string
}

// APIRequest - top level API request
type APIRequest struct {
	JobID uuid.UUID  `json:"jobid"`
	Cmd   APICommand `json:"cmd"`
}

// APIResponse - top level API response
type APIResponse struct {
	JobID uuid.UUID `json:"jobid"`
	Data  []byte    `json:"data"`
	Err   string    `json:"err"`
}
