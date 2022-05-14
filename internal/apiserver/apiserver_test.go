package apiserver

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/buger/jsonparser"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/utils"
	"github.com/neurodyne-web-services/nws-sdk-go/services/cloudcontrol/api"
	"github.com/stretchr/testify/assert"
)

func Test_ssh(t *testing.T) {
	port := rand.Intn(portEnd-portStart) + portStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		// SSH
		{"EC2 SSH List", "List", "NWS::EC2::SSHKeypair", []string{domainID, testAcc}},
		{"EC2 SSH Create", "Create", "NWS::EC2::SSHKeypair", []string{sshKeyName, domainID, testAcc, pubkey}},
		{"EC2 SSH Read", "Read", "NWS::EC2::SSHKeypair", []string{domainID, testAcc}},
		{"EC2 SSH Delete", "Delete", "NWS::EC2::SSHKeypair", []string{sshKeyName, domainID, testAcc}},
		{"EC2 SSH Nuke", "Nuke", "NWS::EC2::SSHKeypair", []string{testAcc, domainID}},
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

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			log.Printf("*** Got test response: %v\n", string(res.Body))
		})
		time.Sleep(200 * time.Millisecond)
	}
}

func TestDS_domain(t *testing.T) {
	port := rand.Intn(portEnd-portStart) + portStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 Zone Read", "Read", zoneCommand, []string{testZone}},
		{"EC2 Domain Read", "Read", domCommand, []string{testDomain}},
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

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			log.Printf("*** Got test response: %v\n", string(res.Body))
		})
		time.Sleep(200 * time.Millisecond)
	}
}

func Test_vpc(t *testing.T) {
	port := rand.Intn(portEnd-portStart) + portStart
	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	var id string

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 VPC List", "List", vpcCommand, []string{zoneID, domainID, testAcc}},
		{"EC2 VPC Create", "Create", vpcCommand, []string{vpcName, zoneID, domainID, testAcc, vpcOfferID, vpcCidr4, netDomain}},
		{"EC2 VPC Resolve", "Resolve", vpcCommand, []string{zoneID, domainID, testAcc, vpcName}},
		{"EC2 VPC Read", "Read", vpcCommand, []string{}},
		{"EC2 VPC Delete", "Delete", vpcCommand, []string{}},
		{"EC2 VPC Nuke", "Nuke", vpcCommand, []string{testAcc, zoneID, domainID}},
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

			if d.action == "Read" || d.action == "Delete" {
				req.Params = append(req.Params, id)
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			if d.action == "Resolve" {
				_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					id, err = jsonparser.GetString(value, "id", "id")
					// fmt.Printf("******* ID resolved: %s\n", id)
					assert.NoError(t, err)
				}, "items")
			}
		})
		time.Sleep(200 * time.Millisecond)
	}
}
