package apiserver

import (
	"context"
	"fmt"
	"log"
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
	accAdmin  = "admin"
	domainID  = "7aa21363-90ec-11ec-83a4-0242ac110003"

	//ssh
	sshKeyName = "bku-ssh"
	pubkey     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDRXZk6v4lDkTkuVHnx/Ztuqv6ntlc6ry5cLjRGyRKOuPGyyaWkK5I1Y2/vtsK8FV6VOJ0Hdjz63kCNaNHtTieDq8W8q2yL2OYiUrgb4cQf3nPs185i41twZBEG12sCBGoXoYNoJl0WsysZ4SlHPgXF+W8BaQK8aJZmFc/f2upjgzX5HxTNhPV5e2ttpvGisH/r8jJBlLZclQa4DHyhq1iTJWNz7DJq6jh4VxqagriRYabuDJRPtTYpi8v5t6+jWbggGIqQkliSaSyYzpHBZAn4PHWUZdRME738IOI2Jy831DH0wvJ0KVjBlcvrT3yXc92iQ9z0s6tFpuQrxMVL3J9+3NmLtKf4i8dcJWDospiQBJp8DrWEVybV34tJk2nHPVzJFpYgJW2XqXdDQhUmQP9CH6L57IDi5Z4vyFvDtcgFd5PFCvkqA7s0PAMF7PY6+laN45qQiO02NFWQHPXbdFyxjzhsHAJPWGWCuPJMwk16fdRgnodk+Ut7j4AfYxSlyRk= bku@lap"
)

func Test_ssh(t *testing.T) {
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
		// Domain
		// {"EC2 Domain List", "Read", "NWS::EC2::Domain", []string{"ROOT"}},

		// // SSH
		// {"EC2 SSH List", "List", "NWS::EC2::SSHKeypair", []string{domainID, accAdmin}},
		// {"EC2 SSH Create", "Create", "NWS::EC2::SSHKeypair", []string{sshKeyName, domainID, accAdmin, pubkey}},
		// {"EC2 SSH Read", "Read", "NWS::EC2::SSHKeypair", []string{domainID, accAdmin}},
		// {"EC2 SSH Delete", "Delete", "NWS::EC2::SSHKeypair", []string{sshKeyName, domainID, accAdmin}},
		{"EC2 SSH Nuke", "Nuke", "NWS::EC2::SSHKeypair", []string{accAdmin, domainID}},
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
			assert.NotEmpty(t, res.Body)

			log.Printf("*** Got test response: %v\n", string(res.Body))
		})
		time.Sleep(500 * time.Millisecond)
	}
}

func TestDS_domain(t *testing.T) {
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
		{"EC2 Zone Read", "Read", "NWS::EC2::Zone", []string{"Sandbox-simulator"}},
		{"EC2 Domain Read", "Read", "NWS::EC2::Domain", []string{"ROOT"}},
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
			assert.NotEmpty(t, res.Body)

			log.Printf("*** Got test response: %v\n", string(res.Body))
		})
		time.Sleep(200 * time.Millisecond)
	}
}
