package apiserver

import (
	"fmt"

	oapimw "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"github.com/neurodyne-web-services/api-gateway/internal/apiserver/api"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
)

const (
	TEST_CONFIG_NAME = "test"
	TEST_CONFIG_PATH = "../../configs"
)

func MakeAPIServerMock() (*echo.Echo, error) {

	// Build a logger
	zl, err := logging.MakeDebugLogger()
	if err != nil {
		log.Fatal("Failed to instantiate a logger")
	}

	// Build a global config
	var cfg config.AppConfig

	if err := cfg.AppInit(TEST_CONFIG_NAME, TEST_CONFIG_PATH); err != nil {
		log.Fatal("Config failed" + err.Error())
	}

	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fail.Error500(err.Error())
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Build a Queue Managers for PING and PONG
	pingMgr, err := BuildQueueManger("PING")
	if err != nil {
		log.Fatalf("Failed to create a queue: %v\n", err)
	}

	pongMgr, err := BuildQueueManger("PONG")
	if err != nil {
		log.Fatalf("Failed to create a queue: %v\n", err)
	}

	// Create an instance of our handler which satisfies the generated interface
	cc := MakeAPIServer(&cfg, zl, pingMgr, pongMgr)

	// This is how you set up a basic Echo router
	e := echo.New()

	// Log all requests
	e.Use(middleware.Logger())

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	e.Use(oapimw.OapiRequestValidator(swagger))

	// Instantiate custom validators
	e.Validator = &CustomValidator{Validator: validator.New()}

	// We now register our cc above as the handler for the interface
	api.RegisterHandlers(e, cc)

	return e, nil
}

func runServer(server *echo.Echo, port int) {
	server.Logger.Fatal(server.Start(fmt.Sprintf("0.0.0.0:%d", port)))
	defer server.Close()
}

func getEndpoint(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}
