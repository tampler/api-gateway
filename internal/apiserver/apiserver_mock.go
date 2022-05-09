package apiserver

import (
	"fmt"
	"log"

	aj "github.com/choria-io/asyncjobs"
	oapimw "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/api-gateway/internal/apiserver/api"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/neurodyne-web-services/api-gateway/internal/nats"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
	uuid "github.com/satori/go.uuid"
)

const (
	CONFIG_PATH = "../../configs"
	CONFIG_NAME = "app"
)

func MakeAPIServerMock() (*echo.Echo, error) {

	// Build a global config
	var cfg config.AppConfig

	if err := cfg.AppInit(CONFIG_NAME, CONFIG_PATH); err != nil {
		log.Fatalf("Config failed %s", err.Error())
	}

	// App logger
	logger, _ := logging.MakeLogger(cfg.Log.Verbosity, cfg.Log.Output)
	defer logger.Sync()
	zl := logger.Sugar()

	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fail.Error500(err.Error())
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Connect to NATS
	nc, err := nats.MakeNatsConnect()
	if err != nil {
		log.Fatalf("NATS connect failed %s \n", err.Error())
	}

	// Input queue
	pingClient, err := aj.NewClient(
		aj.NatsConn(nc),
		aj.BindWorkQueue(cfg.Ajc.Ingress.Name),
		aj.ClientConcurrency(cfg.Ajc.Ingress.Concurrency),
		aj.PrometheusListenPort(cfg.Ajc.Ingress.MetricsPort),
		aj.RetryBackoffPolicy(aj.RetryLinearOneMinute))

	if err != nil {
		zl.Fatal(err)
	}

	// Output queue
	pongClient, err := aj.NewClient(
		aj.NatsConn(nc),
		aj.BindWorkQueue(cfg.Ajc.Egress.Name),
		aj.ClientConcurrency(cfg.Ajc.Egress.Concurrency),
		aj.PrometheusListenPort(cfg.Ajc.Egress.MetricsPort),
		aj.RetryBackoffPolicy(aj.RetryLinearOneMinute))

	if err != nil {
		zl.Fatal(err)
	}

	// Create queue routers
	pingRouter := aj.NewTaskRouter()
	pongRouter := aj.NewTaskRouter()

	pingMgr := MakeQueueManager(pingClient, pingRouter)
	pongMgr := MakeQueueManager(pongClient, pongRouter)

	// Create an instance of our handler which satisfies the generated interface
	cc := MakeAPIServer(&cfg, zl, pingMgr, pongMgr)

	// This is how you set up a basic Echo router
	e := echo.New()

	pub := MakePublisher(pongMgr, zl, map[uuid.UUID]Subscriber{})

	pub.AddHandlers(cfg.Ajc.Egress.Topic)

	// Add custom context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := MakeMyContext(c, cfg, &pub, zl)
			return next(cc)
		}
	})

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
