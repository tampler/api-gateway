package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/casdoor/casdoor-go-sdk/auth"
	aj "github.com/choria-io/asyncjobs"
	oapimw "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/neurodyne-web-services/api-gateway/internal/restserver"
	"github.com/neurodyne-web-services/api-gateway/internal/token"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"github.com/neurodyne-web-services/api-gateway/pkg/rest"
	"github.com/neurodyne-web-services/nws-sdk-go/services/natstool"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	CONFIG_PATH = "./configs"
	CONFIG_NAME = "app"
)

func main() {

	// Build a global config
	var cfg config.AppConfig
	var info restserver.UserInfo

	if err := cfg.AppInit(CONFIG_NAME, CONFIG_PATH); err != nil {
		log.Fatalf("Config failed %s", err.Error())
	}

	// App logger
	logger, _ := logging.MakeLogger(cfg.Log.Verbosity, cfg.Log.Output)
	defer logger.Sync()
	zl := logger.Sugar()

	// Connect to NATS
	nc, err := natstool.MakeNatsConnect()
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

	pingMgr := worker.MakeQueueManager(pingClient, pingRouter)
	pongMgr := worker.MakeQueueManager(pongClient, pongRouter)

	// Create an instance of our handler which satisfies the generated interface
	cc := restserver.MakeRestServer(&cfg, zl, pingMgr, pongMgr)

	// Build Swagger API
	swagger, err := rest.GetSwagger()
	if err != nil {
		zl.Fatalf("Error loading swagger spec: %s", err)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// This is how you set up a basic Echo router
	e := echo.New()

	pub := worker.MakePublisher(pongMgr, zl, map[uuid.UUID]worker.Subscriber{})

	pub.AddHandlers(cfg.Ajc.Egress.Topic)

	// Add custom context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := restserver.MakeMyContext(c, cfg, &pub, zl, info)
			return next(cc)
		}
	})

	// Enable metrics middleware
	p := prometheus.NewPrometheus(cfg.Debug.MetricsName, nil)
	p.Use(e)

	if cfg.Http.AuthEnabled {
		// Read PEM file
		pemData, err := ioutil.ReadFile(cfg.Http.PemFile)
		if err != nil {
			zl.Fatalf("Failed to read the CA file: - %s", err)
		}

		// Read Auth config file
		authData, err := ioutil.ReadFile(cfg.Http.AuthFile)
		if err != nil {
			zl.Fatalf("Failed to read the Auth config file: - %s", err)
		}

		// JWT validator
		r := e.Group("/")
		{
			config := middleware.JWTConfig{
				ParseTokenFunc: func(inputToken string, c echo.Context) (interface{}, error) {
					err := token.InitAuth(authData, pemData)
					if err != nil {
						return nil, err
					}

					claims, err := auth.ParseJwtToken(inputToken)
					if err != nil {
						return nil, errors.New("failed to parse token")
					}

					info.ID = claims.Subject

					return claims.AccessToken, nil
				},
			}
			e.Use(middleware.JWTWithConfig(config))
			r.GET("", restricted)
		}
	}

	// Set Request Limiter
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(cfg.Http.MaxRPS))))

	// Set Request Body size limit
	e.Use(middleware.BodyLimit(cfg.Http.BodyLimit))

	// Set Response Timeout
	if cfg.Http.AllowTimeout {
		e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Timeout: time.Duration(cfg.Http.Timeout) * time.Second,
		}))
	}

	if cfg.Http.AllowLogging {

		// logcfg := middleware.LoggerConfig{
		// 	Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
		// 		`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
		// 		`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
		// 		`"bytes_out":${bytes_out}, "protocol":"${protocol}"}` + "\n",
		// }

		logcfg := middleware.LoggerConfig{
			Format: "status = ${status} time = ${time_rfc3339} lat = ${latency_human}",
		}

		// Log all requests
		// e.Use(middleware.Logger())
		e.Use(middleware.LoggerWithConfig(logcfg))
	}

	if cfg.Http.AllowCompress {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level: cfg.Http.CompressLevel,
		}))
	}

	if cfg.Debug.DumpOnError {
		e.Use(middleware.BodyDump(func(c echo.Context, req, rsp []byte) {
			if c.Response().Status != http.StatusCreated {
				// zl.Error("Request failed", zap.String("RESP", string(rsp)), zap.String("REQ", string(req)))
				zl.Error("Request failed: RESP %v, REQ %v", string(rsp), string(req))
			}
		}))
	}

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	e.Use(oapimw.OapiRequestValidator(swagger))

	// Instantiate custom validators
	e.Validator = &restserver.CustomValidator{Validator: validator.New()}

	// We now register our cloudcontrol above as the handler for the interface
	rest.RegisterHandlers(e, cc)

	showDebugInfo(zl.Desugar(), &cfg)

	// And we serve HTTP until the world ends.
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Http.Port)))
}

func restricted(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.String(http.StatusOK, "Welcome "+name+"!")
}

// showDebugInfo - this prints envs to ease deployment and debug
func showDebugInfo(zl *zap.Logger, cfg *config.AppConfig) {
	zl.Info("NATS URL: ", zap.String("NATS_URL", os.Getenv("NATS_URL")))
	zl.Info("Job timeout:", zap.Int("timeout, sec", cfg.Sdk.JobTime))
}
