package token

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/casdoor/casdoor-go-sdk/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

// ValidateToken - checks input token from GRPC
func ValidateToken(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}

	if !valid(md["authorization"]) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	return handler(ctx, req)
}

func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}

	token := strings.TrimPrefix(authorization[0], "Bearer ")

	// If you have more than one client then you will have to update this line.
	return token == "client-x-id"
}
