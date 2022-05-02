package apiserver

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/neurodyne-web-services/nws-sdk-go/services/cloudcontrol/api"
	"github.com/stretchr/testify/assert"
)

const (
	portStart = 8085
	portEnd   = 9085
	domainID  = "7aa21363-90ec-11ec-83a4-0242ac110003"
)

func Test_server(t *testing.T) {
	port := rand.Intn(portEnd-portStart) + portStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server, port)

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		// List
		{"EC2 Domain List", "Read", "NWS::EC2::Domain", []string{"ROOT"}},
		// {"EC2 SSH List", "List", "NWS::EC2::SSH", []string{domainID, "admin"}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			var req api.CloudControlClient

			client, err := api.NewClientWithResponses(getEndpoint(port))
			assert.NoError(t, err)
			assert.NotNil(t, client)

			req.Client = *client
			req.Action = d.action
			req.Command = d.command
			req.Params = d.params

			res, err := req.MakeRequest(context.Background())
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())

			fmt.Printf("*** Got test response: %v\n", string(res.Body))
		})
		time.Sleep(500 * time.Millisecond)
	}
}
