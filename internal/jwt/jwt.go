package jwt

import (
	"encoding/json"
	"errors"

	"github.com/casdoor/casdoor-go-sdk/auth"
)

// Casdoor Authority init function
func InitAuth(authData, pemData []byte) error {
	var cfg auth.AuthConfig

	err := json.Unmarshal(authData, &cfg)
	if err != nil {
		return errors.New("failed to parse auth file")
	}

	auth.InitConfig(cfg.Endpoint, cfg.ClientId, cfg.ClientSecret, string(pemData), cfg.OrganizationName, cfg.ApplicationName)

	return nil
}
