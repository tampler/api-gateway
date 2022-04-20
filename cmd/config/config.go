package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppConfig struct {
}

func ReadConfig() (AppConfig, error) {
	var cfg AppConfig

	viper.SetConfigName("app")
	viper.AddConfigPath("configs")

	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("Failed to read config file")
	}

	return cfg, nil
}
