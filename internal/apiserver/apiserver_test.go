package apiserver

import (
	"fmt"
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

	tmp, err := server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID := string(tmp.Value())

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		// SSH
		{"EC2 SSH List", "List", sshCommand, []string{domainID, testAcc}},
		{"EC2 SSH Create", "Create", sshCommand, []string{sshKeyName, domainID, testAcc, pubkey}},
		{"EC2 SSH Read", "Read", sshCommand, []string{domainID, testAcc}},
		{"EC2 SSH Delete", "Delete", sshCommand, []string{sshKeyName, domainID, testAcc}},
		{"EC2 SSH Nuke", "Nuke", sshCommand, []string{testAcc, domainID}},
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
		})
		time.Sleep(sleepTime * time.Millisecond)
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
		})
		time.Sleep(sleepTime * time.Millisecond)
	}
}

func Test_vpc(t *testing.T) {

	port := rand.Intn(portEnd-portStart) + portStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	var zoneID, domainID, vpcOfferID, vpcID string

	tmp, err := server.kv.Get("zoneID")
	assert.NoError(t, err)

	zoneID = string(tmp.Value())

	tmp, err = server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID = string(tmp.Value())

	tmp, err = server.kv.Get("vpcOfferID")
	assert.NoError(t, err)

	vpcOfferID = string(tmp.Value())

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
				req.Params = append(req.Params, vpcID)
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			// Resolve runtime VPC ID
			if d.action == "Resolve" {
				_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					vpcID, err = jsonparser.GetString(value, "id", "id")
					assert.NoError(t, err)
				}, "items")
				assert.NoError(t, err)
			}
		})
		time.Sleep(sleepTime * time.Millisecond)
	}
}

func Test_net(t *testing.T) {
	port := rand.Intn(portEnd-portStart) + portStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	var zoneID, domainID, netID, netOfferID, vpcID, vpcOfferID string

	tmp, err := server.kv.Get("zoneID")
	assert.NoError(t, err)

	zoneID = string(tmp.Value())

	tmp, err = server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID = string(tmp.Value())

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 Net List", "List", netCommand, []string{zoneID, domainID, testAcc}},
		{"EC2 VPC Create", "Create", vpcCommand, []string{vpcName, zoneID, domainID, testAcc, vpcOfferID, vpcCidr4, netDomain}},
		{"EC2 VPC ID Resolve", "Resolve", vpcCommand, []string{zoneID, domainID, testAcc, vpcName}},
		{"EC2 Net Offer ID Resolve", "Resolve", netOfferCommand, []string{netOffer}},
		{"EC2 Net Create", "Create", netCommand, []string{netName, zoneID, domainID, testAcc, netCidr4, emptyCIDR6}},
		{"EC2 Net Resolve", "Resolve", netCommand, []string{zoneID, domainID, testAcc, netName}},
		{"EC2 Net Read", "Read", netCommand, []string{}},
		{"EC2 Net Delete", "Delete", netCommand, []string{}},
		{"EC2 Net Nuke", "Nuke", netCommand, []string{testAcc, zoneID, domainID}},
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

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Read" || d.action == "Delete" {
				req.Params = append(req.Params, netID)
			}

			if d.action == "Create" {
				req.Params = append(req.Params, netOfferID, netDomain, vpcID)
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			// Resolve for diff IDs may be in diff test cases
			if d.action == "Resolve" {

				// parse  Network Offer ID
				if d.command == netOfferCommand {

					netOfferID, err = jsonparser.GetString(data, "id")
					assert.NoError(t, err)
				}

				if d.command == vpcCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
						vpcID, err = jsonparser.GetString(value, "id", "id")
						assert.NoError(t, err)
					}, "items")
					assert.NoError(t, err)
				}

				if d.command == netCommand {
					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
						netID, err = jsonparser.GetString(value, "id", "id")
						assert.NoError(t, err)
					}, "items")
					assert.NoError(t, err)
				}
			}
		})
		time.Sleep(sleepTime * time.Millisecond)
	}
}

func Test_tmpl(t *testing.T) {
	port := rand.Intn(portEnd-portStart) + portStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	var zoneID, domainID, osTypeID, tmplID string

	tmp, err := server.kv.Get("zoneID")
	assert.NoError(t, err)

	zoneID = string(tmp.Value())

	tmp, err = server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID = string(tmp.Value())

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 Tmpl List", "List", tmplCommand, []string{zoneID, domainID, testAcc, tmplFilter}},
		{"EC2 OS Offer Resolve", "Resolve", osOfferCommand, []string{ostype}},
		{"EC2 Tmpl Create", "Create", tmplCommand, []string{tmplName, zoneID, domainID, testAcc}},
		{"EC2 Tmpl Resolve", "Resolve", tmplCommand, []string{zoneID, domainID, testAcc, tmplFilter, tmplName}},
		{"EC2 Tmpl Read", "Read", tmplCommand, []string{}},
		{"EC2 Tmpl Delete", "Delete", tmplCommand, []string{}},
		{"EC2 Tmpl Nuke", "Nuke", tmplCommand, []string{testAcc, tmplFilter, zoneID, domainID}},
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

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Create" {
				req.Params = append(req.Params, osTypeID, tmplURL)
			}

			if d.action == "Read" {
				req.Params = append(req.Params, tmplID, tmplFilter)
			}

			if d.action == "Delete" {
				req.Params = append(req.Params, tmplID, zoneID)
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			if d.action == "Resolve" {

				if d.command == osOfferCommand {
					osTypeID, err = jsonparser.GetString(data, "id")
					assert.NoError(t, err)
				}

				if d.command == tmplCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

						if string(value) != jsonparser.Null.String() {
							tmplID, err = jsonparser.GetString(value, "id", "id")
							assert.NoError(t, err)
						}

						assert.NoError(t, err)
					}, "items")

					assert.NoError(t, err)
				}
			}
		})
		time.Sleep(sleepTime * time.Millisecond)
	}
}

func Test_inst(t *testing.T) {
	port := rand.Intn(portEnd-portStart) + portStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	var zoneID, domainID, osTypeID, tmplID, vpcOfferID string

	tmp, err := server.kv.Get("zoneID")
	assert.NoError(t, err)

	zoneID = string(tmp.Value())

	tmp, err = server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID = string(tmp.Value())

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		// {"EC2 Inst List", "List", instCommand, []string{zoneID, domainID, testAcc}},
		// {"EC2 SSH Create", "Create", sshCommand, []string{sshKeyName, domainID, testAcc, pubkey}},
		{"EC2 VPC Create", "Create", vpcCommand, []string{vpcName, zoneID, domainID, testAcc, vpcOfferID, vpcCidr4, netDomain}},
		{"EC2 Net Create", "Create", netCommand, []string{netName, zoneID, domainID, testAcc, netCidr4, emptyCIDR6}},
		// {"EC2 Tmpl Create", "Create", tmplCommand, []string{tmplName, zoneID, domainID, testAcc}},
		// {"EC2 Inst Create", "Create", instCommand, []string{instName, zoneID, domainID, testAcc, tmplID}},
		// {"EC2 Inst Resolve", "Resolve", tmplCommand, []string{zoneID, domainID, testAcc,  tmplName}},
		// {"EC2 Inst Read", "Read", tmplCommand, []string{}},
		// {"EC2 Inst Delete", "Delete", tmplCommand, []string{}},
		// {"EC2 Inst Nuke", "Nuke", tmplCommand, []string{testAcc,  zoneID, domainID}},
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

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Create" {
				req.Params = append(req.Params, osTypeID, tmplURL)
			}

			if d.action == "Read" {
				req.Params = append(req.Params, tmplID, tmplFilter)
			}

			if d.action == "Delete" {
				req.Params = append(req.Params, tmplID, zoneID)
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			if d.action == "Resolve" {

				if d.command == osOfferCommand {
					osTypeID, err = jsonparser.GetString(data, "id")
					assert.NoError(t, err)
				}

				if d.command == tmplCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

						if string(value) != jsonparser.Null.String() {
							tmplID, err = jsonparser.GetString(value, "id", "id")
							assert.NoError(t, err)
						}

						assert.NoError(t, err)
					}, "items")

					assert.NoError(t, err)
				}
			}
		})
		time.Sleep(sleepTime * time.Millisecond)
	}
}

func Test_offerings(t *testing.T) {
	port := rand.Intn(portEnd-portStart) + portStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoError(t, err)

	go runServer(server.echo, port)

	var zoneID, domainID, vpcOfferID string

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 Zone ID", "Resolve", zoneCommand, []string{testZone}},
		{"EC2 Domain ID", "Resolve", domCommand, []string{testDomain}},
		{"EC2 VPC Offer ID", "Resolve", vpcOfferCommand, []string{vpcOffer}},
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

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			// Store Zone ID
			if d.command == zoneCommand {
				zoneID, err = jsonparser.GetString(data, "id")
				assert.NoError(t, err)
				assert.NotEmpty(t, zoneID)
				server.kv.Put("zoneID", []byte(zoneID))
			}

			// Store Domain ID
			if d.command == domCommand {
				domainID, err = jsonparser.GetString(data, "id")
				assert.NoError(t, err)
				assert.NotEmpty(t, domainID)
				server.kv.Put("domainID", []byte(domainID))
			}

			// Store VPC Offer ID
			if d.command == vpcOfferCommand {
				vpcOfferID, err = jsonparser.GetString(data, "id")
				assert.NoError(t, err)
				assert.NotEmpty(t, vpcOfferID)
				server.kv.Put("vpcOfferID", []byte(vpcOfferID))
			}
		})
		time.Sleep(sleepTime * time.Millisecond)
	}
}
