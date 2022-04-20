package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/casdoor/casdoor-go-sdk/auth"
	oapimw "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/neurodyne-web-services/api-gateway/internal/cloudcontrol"
	"github.com/neurodyne-web-services/api-gateway/internal/cloudcontrol/api"
	njwt "github.com/neurodyne-web-services/api-gateway/internal/jwt"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const (
	CONFIG_PATH = "./configs"
	CONFIG_NAME = "app"
)

func main() {

	// Build logger
	zl, err := logging.MakeDebugLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to instantiate a logger")
		os.Exit(1)
	}

	// Read global config
	viper.SetConfigName(CONFIG_NAME)
	viper.AddConfigPath(CONFIG_PATH)

	if err := viper.ReadInConfig(); err != nil {
		zl.Fatal("Failed to read config file", zap.String("config name", CONFIG_NAME), zap.String("config path", CONFIG_PATH))
	}

	// App config
	port := viper.GetInt("http.port")
	pemFile := viper.GetString("http.pem_file")
	authFile := viper.GetString("http.auth_file")
	maxRPS := viper.GetInt("http.max_rps")
	bodyLimit := viper.GetString("http.body_limit")
	allowTimeout := viper.GetBool("http.allow_timeout")
	timeout := viper.GetInt("http.timeout")
	allowCompress := viper.GetBool("http.allow_compress")
	compressLevel := viper.GetInt("http.compress_level")
	dumpOnError := viper.GetBool("debug.dump_on_error")
	metricsName := viper.GetString("debug.metrics_name")

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Create an instance of our handler which satisfies the generated interface
	cc := cloudcontrol.MakeAPIServer()

	// This is how you set up a basic Echo router
	e := echo.New()

	// Enable metrics middleware
	p := prometheus.NewPrometheus(metricsName, nil)
	p.Use(e)

	// Read PEM file
	pemData, err := ioutil.ReadFile(pemFile)
	if err != nil {
		log.Fatalf("Failed to read the CA file: - %s", err)
	}

	// Read Auth config file
	authData, err := ioutil.ReadFile(authFile)
	if err != nil {
		log.Fatalf("Failed to read the Auth config file: - %s", err)
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

	// Set Request Limiter
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(maxRPS))))

	// Set Request Body size limit
	e.Use(middleware.BodyLimit(bodyLimit))

	// Set Response Timeout
	if allowTimeout {
		e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
			Timeout: time.Duration(timeout) * time.Second,
		}))
	}

	// cfg := middleware.LoggerConfig{
	// 	Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
	// 		`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
	// 		`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
	// 		`"bytes_out":${bytes_out}, "protocol":"${protocol}"}` + "\n",
	// }
	// cfg := middleware.LoggerConfig{
	// 	Format: "status = ${status} time = ${time_rfc3339} lat = ${latency_human} \n",
	// }

	// // Log all requests
	// e.Use(middleware.LoggerWithConfig(cfg))
	e.Use(middleware.Logger())

	if allowCompress {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level: compressLevel,
		}))
	}

	if dumpOnError {
		e.Use(middleware.BodyDump(func(c echo.Context, req, rsp []byte) {
			if c.Response().Status != http.StatusCreated {
				zl.Error("Request failed", zap.String("RESP", string(rsp)), zap.String("REQ", string(req)))
			}
		}))
	}

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	e.Use(oapimw.OapiRequestValidator(swagger))

	// Instantiate custom validators
	e.Validator = &cloudcontrol.CustomValidator{Validator: validator.New()}

	// We now register our cloudcontrol above as the handler for the interface
	api.RegisterHandlers(e, cc)

	// And we serve HTTP until the world ends.
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", port)))
}

func restricted(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.String(http.StatusOK, "Welcome "+name+"!")
}
