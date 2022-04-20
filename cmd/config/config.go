package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Port          int
	PemFile       string
	AuthFile      string
	MaxRPS        int
	BodyLimit     string
	AllowTimeout  bool
	Timeout       int
	AllowCompress bool
	CompressLevel int
	DumpOnError   bool
	MetricsName   string
}

func (cfg *AppConfig) AppInit() error {

	viper.SetConfigName("app")
	viper.AddConfigPath("configs")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Failed to read config file")
	}

	cfg.Port = viper.GetInt("http.port")
	cfg.PemFile = viper.GetString("http.pem_file")
	cfg.AuthFile = viper.GetString("http.auth_file")
	cfg.MaxRPS = viper.GetInt("http.max_rps")
	cfg.BodyLimit = viper.GetString("http.body_limit")
	cfg.AllowTimeout = viper.GetBool("http.allow_timeout")
	cfg.Timeout = viper.GetInt("http.timeout")
	cfg.AllowCompress = viper.GetBool("http.allow_compress")
	cfg.CompressLevel = viper.GetInt("http.compress_level")
	cfg.DumpOnError = viper.GetBool("debug.dump_on_error")
	cfg.MetricsName = viper.GetString("debug.metrics_name")

	fmt.Printf("CA File: %s", cfg.PemFile)

	return nil
}
