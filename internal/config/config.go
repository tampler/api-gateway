package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// httpConfig - config for Http and JWT server
type httpConfig struct {
	Port          int
	AuthEnabled   bool
	PemFile       string
	AuthFile      string
	MaxRPS        int
	BodyLimit     string
	AllowTimeout  bool
	Timeout       int
	AllowCompress bool
	AllowLogging  bool
	CompressLevel int
}

// debugConfig - config for debugging
type debugConfig struct {
	DumpOnError bool
	MetricsName string
}

type queueConfig struct {
	Name        string
	Topic       string
	Concurrency int
	MetricsPort int
}

type ajcConfig struct {
	Timeout int
	Ingress queueConfig
	Egress  queueConfig
}

type logConfig struct {
	Verbosity string
	Output    string
}

type sdkConfig struct {
	JobTime int
	Bucket  string
}

// AppConfig - top level config
type AppConfig struct {
	Log   logConfig
	Debug debugConfig
	Http  httpConfig
	Ajc   ajcConfig
	Sdk   sdkConfig
}

// AppInit - reads config file
func (cfg *AppConfig) AppInit(name, path string) error {

	viper.SetConfigName(name)
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Failed to read config file")
	}

	// auth
	cfg.Http.AuthEnabled = viper.GetBool("auth.auth_enabled")

	// http setup
	cfg.Http.Port = viper.GetInt("http.port")
	cfg.Http.PemFile = viper.GetString("http.pem_file")
	cfg.Http.AuthFile = viper.GetString("http.auth_file")
	cfg.Http.MaxRPS = viper.GetInt("http.max_rps")
	cfg.Http.BodyLimit = viper.GetString("http.body_limit")
	cfg.Http.AllowTimeout = viper.GetBool("http.allow_timeout")
	cfg.Http.Timeout = viper.GetInt("http.timeout")
	cfg.Http.AllowCompress = viper.GetBool("http.allow_compress")
	cfg.Http.AllowLogging = viper.GetBool("http.allow_logging")
	cfg.Http.CompressLevel = viper.GetInt("http.compress_level")

	// debug
	cfg.Debug.DumpOnError = viper.GetBool("debug.dump_on_error")
	cfg.Debug.MetricsName = viper.GetString("debug.metrics_name")

	// log
	cfg.Log.Verbosity = viper.GetString("debug.log_verbosity")
	cfg.Log.Output = viper.GetString("debug.log_output")

	// Ping queue
	cfg.Ajc.Ingress.Name = viper.GetString("ajc.ingress.name")
	cfg.Ajc.Ingress.Topic = viper.GetString("ajc.topic")
	cfg.Ajc.Ingress.Concurrency = viper.GetInt("ajc.concurrency")
	cfg.Ajc.Ingress.MetricsPort = viper.GetInt("ajc.ingress.metrics_port")

	// Pong queue
	cfg.Ajc.Egress.Name = viper.GetString("ajc.egress.name")
	cfg.Ajc.Egress.Topic = viper.GetString("ajc.topic")
	cfg.Ajc.Egress.Concurrency = viper.GetInt("ajc.concurrency")
	cfg.Ajc.Egress.MetricsPort = viper.GetInt("ajc.egress.metrics_port")

	// SDK
	cfg.Sdk.JobTime = viper.GetInt("sdk.job_time_sec")
	cfg.Ajc.Timeout = viper.GetInt("ajc.task_deadline_min")
	cfg.Sdk.Bucket = viper.GetString("sdk.kv_bucket_name")

	return nil
}
