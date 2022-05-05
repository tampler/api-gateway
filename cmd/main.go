package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/casdoor/casdoor-go-sdk/auth"
	oapimw "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/neurodyne-web-services/api-gateway/cmd/config"
	"github.com/neurodyne-web-services/api-gateway/internal/apiserver"
	"github.com/neurodyne-web-services/api-gateway/internal/apiserver/api"
	njwt "github.com/neurodyne-web-services/api-gateway/internal/jwt"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/time/rate"
)

const (
	CONFIG_PATH = "./configs"
	CONFIG_NAME = "app"
	logOut      = "console" // json/console

)

func main() {

	// SDK logger
	logger, _ := logging.MakeLogger("debug", logOut)
	defer logger.Sync()
	zl := logger.Sugar()

	// Build a global config
	var cfg config.AppConfig

	if err := cfg.AppInit(CONFIG_NAME, CONFIG_PATH); err != nil {
		zl.Fatal("Config failed %s", err.Error())
	}

	// Build a Queue Managers for PING and PONG
	pingMgr, err := apiserver.BuildQueueManger("PING", zl)
	if err != nil {
		zl.Fatalf("Failed to create a queue: %v\n", err)
	}

	pongMgr, err := apiserver.BuildQueueManger("PONG", zl)
	if err != nil {
		zl.Fatalf("Failed to create a queue: %v\n", err)
	}

	// Create an instance of our handler which satisfies the generated interface
	cc := apiserver.MakeAPIServer(&cfg, zl, pingMgr, pongMgr)

	// Build Swagger API
	swagger, err := api.GetSwagger()
	if err != nil {
		zl.Fatalf("Error loading swagger spec\n: %s", err)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// This is how you set up a basic Echo router
	e := echo.New()

	pub := apiserver.MakePublisher(pongMgr, zl, map[uuid.UUID]apiserver.Subscriber{})

	pub.AddHandlers()

	// Add custom context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := apiserver.MakeMyContext(c, &pub, zl)
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
				ParseTokenFunc: func(token string, c echo.Context) (interface{}, error) {
					err := njwt.InitAuth(authData, pemData)
					if err != nil {
						return nil, err
					}

					claims, err := auth.ParseJwtToken(token)
					if err != nil {
						return nil, errors.New("failed to parse token")
					}

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

	// cfg := middleware.LoggerConfig{
	// 	Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
	// 		`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
	// 		`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
	// 		`"bytes_out":${bytes_out}, "protocol":"${protocol}"}` + "\n",
	// }
	logcfg := middleware.LoggerConfig{
		Format: "status = ${status} time = ${time_rfc3339} lat = ${latency_human} \n",
	}

	// Log all requests
	e.Use(middleware.LoggerWithConfig(logcfg))
	// e.Use(middleware.Logger())

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
	e.Validator = &apiserver.CustomValidator{Validator: validator.New()}

	// We now register our cloudcontrol above as the handler for the interface
	api.RegisterHandlers(e, cc)

	// And we serve HTTP until the world ends.
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Http.Port)))
}

func restricted(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.String(http.StatusOK, "Welcome "+name+"!")
}
