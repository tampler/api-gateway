package jwt

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/casdoor/casdoor-go-sdk/auth"
	"github.com/stretchr/testify/assert"
)

const (
	AUTH_FILE = "../../../configs/jwt/app.json"
	PEM_FILE  = "/tmp/token_jwt_key.pem"
)

func TestAuth_token(t *testing.T) {

	fmt.Printf("Getting the Token... \n")

	// Read PEM file
	pemData, err := ioutil.ReadFile(PEM_FILE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read the CA file: - %s", err)
		os.Exit(1)
	}

	// Read Auth config file
	authData, err := ioutil.ReadFile(AUTH_FILE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read the Auth config file: - %s", err)
		os.Exit(1)
	}

	err = InitAuth(authData, pemData)
	assert.NoError(t, err)

	user, err := auth.GetUser("bku")

	fmt.Printf("User: %v \n", user)

	_, err = auth.GetOAuthToken("7561ce5aa2bee12760ba", "app-sdk-back")
	assert.NoError(t, err)

	// fmt.Printf("Token: %s", token.AccessToken)
}
