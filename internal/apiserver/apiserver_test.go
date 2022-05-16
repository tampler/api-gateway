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

func Test_sess(t *testing.T) {
	// os.Setenv("NATS_URL", "192.168.1.93:42377")
	// os.Setenv("NATS_USER", "local")
	// os.Setenv("NATS_PASS", "vsO2TcFwkmQ2p3eiOl3HcD7NVWEjvRI4")

	port := rand.Intn(portEnd-portStart) + portStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	var zoneID, domainID, account string

	account = "admin"

	tmp, err := server.kv.Get("zoneID")
	assert.NoError(t, err)

	zoneID = string(tmp.Value())

	tmp, err = server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID = string(tmp.Value())

	now := time.Now().Format(timeFormat)

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"Session Create", "Create", sessCommand, []string{userID, testZone, zoneID, testDomain, domainID, account, now}},
		{"Session Read", "Read", sessCommand, []string{userID}},
		{"Session Delete", "Delete", sessCommand, []string{userID}},
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

			fmt.Println(string(data))

			if req.Action == "Read" {

				readZone, err := jsonparser.GetString(data, "zone", "name")
				assert.NoError(t, err)
				assert.Equal(t, testZone, readZone)

				readZoneID, err := jsonparser.GetString(data, "zone", "id")
				assert.NoError(t, err)
				assert.Equal(t, zoneID, readZoneID)

				readDomain, err := jsonparser.GetString(data, "domain", "name")
				assert.NoError(t, err)
				assert.Equal(t, testDomain, readDomain)

				readDomainID, err := jsonparser.GetString(data, "domain", "id")
				assert.NoError(t, err)
				assert.Equal(t, domainID, readDomainID)

				readAccount, err := jsonparser.GetString(data, "account")
				assert.NoError(t, err)
				assert.Equal(t, account, readAccount)
			}
		})
		time.Sleep(sleepTime * time.Millisecond)
	}
}

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

	var zoneID, domainID, netID, vpcID, vpcOfferID, netOfferID string

	tmp, err := server.kv.Get("zoneID")
	assert.NoError(t, err)

	zoneID = string(tmp.Value())

	tmp, err = server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID = string(tmp.Value())

	tmp, err = server.kv.Get("vpcOfferID")
	assert.NoError(t, err)

	vpcOfferID = string(tmp.Value())

	tmp, err = server.kv.Get("netOfferID")
	assert.NoError(t, err)

	netOfferID = string(tmp.Value())

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 Net List", "List", netCommand, []string{zoneID, domainID, testAcc}},
		{"EC2 VPC Create", "Create", vpcCommand, []string{vpcName, zoneID, domainID, testAcc, vpcOfferID, vpcCidr4, netDomain}},
		{"EC2 VPC ID Resolve", "Resolve", vpcCommand, []string{zoneID, domainID, testAcc, vpcName}},
		{"EC2 Net Create", "Create", netCommand, []string{netName, zoneID, domainID, testAcc, netCidr4, emptyCIDR6, netOfferID, netDomain}},
		{"EC2 Net Resolve", "Resolve", netCommand, []string{zoneID, domainID, testAcc, netName}},
		{"EC2 Net Read", "Read", netCommand, []string{}},
		{"EC2 Net Delete", "Delete", netCommand, []string{}},
		{"EC2 Net Nuke", "Nuke", netCommand, []string{testAcc, zoneID, domainID}},
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

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Read" || d.action == "Delete" {
				req.Params = append(req.Params, netID)
			}

			if d.action == "Create" {
				req.Params = append(req.Params, vpcID)
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			// Resolve for diff IDs may be in diff test cases
			if d.action == "Resolve" {

				// Resolves VPC ID
				if d.command == vpcCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
						vpcID, err = jsonparser.GetString(value, "id", "id")
						assert.NoError(t, err)
					}, "items")
					assert.NoError(t, err)
				}

				// Resolves NET ID
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

	var zoneID, domainID, osOfferID, tmplID string

	tmp, err := server.kv.Get("zoneID")
	assert.NoError(t, err)

	zoneID = string(tmp.Value())

	tmp, err = server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID = string(tmp.Value())

	tmp, err = server.kv.Get("osOfferID")
	assert.NoError(t, err)

	osOfferID = string(tmp.Value())

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 Tmpl List", "List", tmplCommand, []string{zoneID, domainID, testAcc, tmplFilter}},
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
				req.Params = append(req.Params, osOfferID, tmplURL)
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
					osOfferID, err = jsonparser.GetString(data, "id")
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

	var zoneID, domainID, osOfferID, vpcOfferID, netOfferID, instOfferID string
	var vpcID, netID, tmplID, instID string

	fmt.Println(instOfferID)

	tmp, err := server.kv.Get("zoneID")
	assert.NoError(t, err)

	zoneID = string(tmp.Value())

	tmp, err = server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID = string(tmp.Value())

	tmp, err = server.kv.Get("vpcOfferID")
	assert.NoError(t, err)

	vpcOfferID = string(tmp.Value())

	tmp, err = server.kv.Get("netOfferID")
	assert.NoError(t, err)

	netOfferID = string(tmp.Value())

	tmp, err = server.kv.Get("osOfferID")
	assert.NoError(t, err)

	osOfferID = string(tmp.Value())

	tmp, err = server.kv.Get("instOfferID")
	assert.NoError(t, err)

	instOfferID = string(tmp.Value())

	fmt.Println(vpcOfferID)
	fmt.Println(netOfferID)

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 Inst List", "List", instCommand, []string{zoneID, domainID, testAcc}},
		{"EC2 SSH Create", "Create", sshCommand, []string{sshKeyName, domainID, testAcc, pubkey}},
		{"EC2 VPC Create", "Create", vpcCommand, []string{vpcName, zoneID, domainID, testAcc, vpcOfferID, vpcCidr4, netDomain}},
		{"EC2 VPC Resolve", "Resolve", vpcCommand, []string{zoneID, domainID, testAcc, vpcName}},
		{"EC2 Net Create", "Create", netCommand, []string{netName, zoneID, domainID, testAcc, netCidr4, emptyCIDR6, netOfferID, netDomain}},
		{"EC2 Net Resolve", "Resolve", netCommand, []string{zoneID, domainID, testAcc, netName}},
		{"EC2 Tmpl Create", "Create", tmplCommand, []string{tmplName, zoneID, domainID, testAcc}},
		{"EC2 Tmpl Resolve", "Resolve", tmplCommand, []string{zoneID, domainID, testAcc, tmplFilter, tmplName}},
		{"EC2 Inst Create", "Create", instCommand, []string{instName, zoneID, domainID, testAcc}},
		{"EC2 Inst Resolve", "Resolve", instCommand, []string{zoneID, domainID, testAcc}},
		{"EC2 Inst Read", "Read", instCommand, []string{}},
		{"EC2 Inst Delete", "Delete", instCommand, []string{}},
		{"EC2 Inst Nuke", "Nuke", instCommand, []string{testAcc, zoneID, domainID}},
		{"EC2 Tmpl Nuke", "Nuke", tmplCommand, []string{testAcc, tmplFilter, zoneID, domainID}},
		{"EC2 Net Nuke", "Nuke", netCommand, []string{testAcc, zoneID, domainID}},
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

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Create" {

				if d.command == netCommand {
					req.Params = append(req.Params, vpcID)
				}

				if d.command == tmplCommand {
					req.Params = append(req.Params, osOfferID, tmplURL)
				}

				if d.command == instCommand {
					req.Params = append(req.Params, tmplID, instOfferID, sshKeyName, fmt.Sprint(diskSizeGB), fmt.Sprintf("net::%s", netID))
				}
			}

			if d.action == "Resolve" {
				if d.command == instCommand {
					req.Params = append(req.Params, tmplID, instName)
				}
			}

			if d.action == "Read" {
				if d.command == tmplCommand {
					req.Params = append(req.Params, tmplID, tmplFilter)
				}
				if d.command == instCommand {
					req.Params = append(req.Params, instID)
				}
			}

			if d.action == "Delete" {
				req.Params = append(req.Params, instID)
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			if d.action == "Resolve" {

				// Resolve VPC ID
				if d.command == vpcCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
						vpcID, err = jsonparser.GetString(value, "id", "id")
						assert.NoError(t, err)
					}, "items")
					assert.NoError(t, err)
				}

				// Resolve Net ID
				if d.command == netCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
						netID, err = jsonparser.GetString(value, "id", "id")
						assert.NoError(t, err)
					}, "items")
					assert.NoError(t, err)
				}

				// Resolve Templ ID
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

				// Resolve Inst ID
				if d.command == instCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

						if string(value) != jsonparser.Null.String() {
							instID, err = jsonparser.GetString(value, "id", "id")
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

	var zoneID, domainID, vpcOfferID, netOfferID, osOfferID, instOfferID string

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 Zone ID", "Resolve", zoneCommand, []string{testZone}},
		{"EC2 Domain ID", "Resolve", domCommand, []string{testDomain}},
		{"EC2 VPC Offer ID", "Resolve", vpcOfferCommand, []string{vpcOffer}},
		{"EC2 Net Offer ID", "Resolve", netOfferCommand, []string{netOffer}},
		{"EC2 OS Offer ID", "Resolve", osOfferCommand, []string{osOffer}},
		{"EC2 Inst Offer ID", "Resolve", instOfferCommand, []string{instOffer}},
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
				_, err = server.kv.Put("zoneID", []byte(zoneID))
				assert.NoError(t, err)
			}

			// Store Domain ID
			if d.command == domCommand {
				domainID, err = jsonparser.GetString(data, "id")
				assert.NoError(t, err)
				assert.NotEmpty(t, domainID)
				_, err = server.kv.Put("domainID", []byte(domainID))
				assert.NoError(t, err)
			}

			// Store VPC Offer ID
			if d.command == vpcOfferCommand {
				vpcOfferID, err = jsonparser.GetString(data, "id")
				assert.NoError(t, err)
				assert.NotEmpty(t, vpcOfferID)
				_, err = server.kv.Put("vpcOfferID", []byte(vpcOfferID))
				assert.NoError(t, err)
			}

			// Store Net Offer ID
			if d.command == netOfferCommand {
				netOfferID, err = jsonparser.GetString(data, "id")
				assert.NoError(t, err)
				assert.NotEmpty(t, netOfferID)
				_, err = server.kv.Put("netOfferID", []byte(netOfferID))
				assert.NoError(t, err)
			}

			// Store OS Offer ID
			if d.command == osOfferCommand {
				osOfferID, err = jsonparser.GetString(data, "id")
				assert.NoError(t, err)
				assert.NotEmpty(t, osOfferID)
				_, err = server.kv.Put("osOfferID", []byte(osOfferID))
				assert.NoError(t, err)
			}

			// Store Inst Offer ID
			if d.command == instOfferCommand {
				instOfferID, err = jsonparser.GetString(data, "id")
				assert.NoError(t, err)
				assert.NotEmpty(t, instOfferID)
				_, err = server.kv.Put("instOfferID", []byte(instOfferID))
				assert.NoError(t, err)
			}
		})
		time.Sleep(sleepTime * time.Millisecond)
	}
}
