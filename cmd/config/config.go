package config

import (
	"fmt"

	"github.com/spf13/viper"
)

//NatsConfig - config for Nats client
type NatsConfig struct {
	Server  string
	Topic   string
	Timeout int
}

//HttpConfig - config for Http and JWT server
type HttpConfig struct {
	Port          int
	PemFile       string
	AuthFile      string
	MaxRPS        int
	BodyLimit     string
	AllowTimeout  bool
	Timeout       int
	AllowCompress bool
	CompressLevel int
}

//DebugConfig - config for debugging
type DebugConfig struct {
	DumpOnError bool
	MetricsName string
}

//AppConfig - top level config
type AppConfig struct {
	Http  HttpConfig
	Nats  NatsConfig
	Debug DebugConfig
}

//AppInit - reads config file
func (cfg *AppConfig) AppInit(name, path string) error {

	viper.SetConfigName(name)
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Failed to read config file")
	}

	// http setup
	cfg.Http.Port = viper.GetInt("http.port")
	cfg.Http.PemFile = viper.GetString("http.pem_file")
	cfg.Http.AuthFile = viper.GetString("http.auth_file")
	cfg.Http.MaxRPS = viper.GetInt("http.max_rps")
	cfg.Http.BodyLimit = viper.GetString("http.body_limit")
	cfg.Http.AllowTimeout = viper.GetBool("http.allow_timeout")
	cfg.Http.Timeout = viper.GetInt("http.timeout")
	cfg.Http.AllowCompress = viper.GetBool("http.allow_compress")
	cfg.Http.CompressLevel = viper.GetInt("http.compress_level")

	// debug
	cfg.Debug.DumpOnError = viper.GetBool("debug.dump_on_error")
	cfg.Debug.MetricsName = viper.GetString("debug.metrics_name")

	// nats
	cfg.Nats.Server = viper.GetString("nats.server")
	cfg.Nats.Topic = viper.GetString("nats.topic")
	cfg.Nats.Timeout = viper.GetInt("nats.timeout")

	return nil
}
