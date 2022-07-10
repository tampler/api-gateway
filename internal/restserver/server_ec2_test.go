package restserver

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/buger/jsonparser"
	"github.com/neurodyne-web-services/api-gateway/internal/common"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/utils"
	cc "github.com/neurodyne-web-services/nws-sdk-go/services/cloudcontrol"
	"github.com/stretchr/testify/assert"
)

func Test_ssh(t *testing.T) {
	// os.Setenv("TF_ACC", "1")
	// os.Setenv("NATS_URL", "192.168.1.93:37487")
	// os.Setenv("NATS_USER", "local")
	// os.Setenv("NATS_PASS", "")
	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	tmp, err := server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID := string(tmp.Value())

	var sshKeyID string

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 SSH List", "List", sshCommand, []string{domainID, common.TestAcc}},
		{"EC2 SSH Create", "Create", sshCommand, []string{common.SshKeyName, domainID, common.TestAcc, common.Pubkey}},
		{"EC2 SSH Resolve", "Resolve", sshCommand, []string{domainID, common.TestAcc, common.SshKeyName}},
		{"EC2 SSH Read", "Read", sshCommand, []string{}},
		{"EC2 SSH Delete", "Delete", sshCommand, []string{common.SshKeyName, domainID, common.TestAcc}},
		{"EC2 SSH Nuke", "Nuke", sshCommand, []string{common.TestAcc, domainID}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			req, err := cc.MakePlainClient(getEndpoint(port))
			assert.NoError(t, err)

			req.Cmd.Action = d.action
			req.Cmd.Command = d.command
			req.Cmd.Params = d.params

			if d.action == "Read" || d.action == "Delete" {
				req.Cmd.Params = append(req.Cmd.Params, sshKeyID)
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))

			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			// Resolve runtime VPC ID
			if d.action == "Resolve" {
				sshKeyID, err = jsonparser.GetString(data, "id", "id")
				assert.NoError(t, err)
			}
		})
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}

func TestDS_domain(t *testing.T) {
	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

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
		{"EC2 Zone Read", "Read", zoneCommand, []string{common.TestZone}},
		{"EC2 Domain Read", "Read", domCommand, []string{common.TestDomain}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			req, err := cc.MakePlainClient(getEndpoint(port))
			assert.NoError(t, err)

			req.Cmd.Action = d.action
			req.Cmd.Command = d.command
			req.Cmd.Params = d.params

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)
		})
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}

func Test_vpc(t *testing.T) {

	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

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
		{"EC2 VPC List", "List", vpcCommand, []string{zoneID, domainID, common.TestAcc}},
		{"EC2 VPC Create", "Create", vpcCommand, []string{common.VpcName, zoneID, domainID, common.TestAcc, vpcOfferID, common.VpcCidr4, common.NetDomain}},
		{"EC2 VPC Resolve", "Resolve", vpcCommand, []string{zoneID, domainID, common.TestAcc, common.VpcName}},
		{"EC2 VPC Read", "Read", vpcCommand, []string{}},
		{"EC2 VPC Delete", "Delete", vpcCommand, []string{}},
		{"EC2 VPC Nuke", "Nuke", vpcCommand, []string{common.TestAcc, zoneID, domainID}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			req, err := cc.MakePlainClient(getEndpoint(port))
			assert.NoError(t, err)

			req.Cmd.Action = d.action
			req.Cmd.Command = d.command
			req.Cmd.Params = d.params

			if d.action == "Read" || d.action == "Delete" {
				req.Cmd.Params = append(req.Cmd.Params, vpcID)
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
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}

func Test_net(t *testing.T) {
	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

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
		{"EC2 Net List", "List", netCommand, []string{zoneID, domainID, common.TestAcc}},
		{"EC2 VPC Create", "Create", vpcCommand, []string{common.VpcName, zoneID, domainID, common.TestAcc, vpcOfferID, common.VpcCidr4, common.NetDomain}},
		{"EC2 VPC ID Resolve", "Resolve", vpcCommand, []string{zoneID, domainID, common.TestAcc, common.VpcName}},
		{"EC2 Net Create", "Create", netCommand, []string{common.NetName, zoneID, domainID, common.TestAcc, common.NetCidr4, common.EmptyCIDR6, netOfferID, common.NetDomain}},
		{"EC2 Net Resolve", "Resolve", netCommand, []string{zoneID, domainID, common.TestAcc, common.NetName}},
		{"EC2 Net Read", "Read", netCommand, []string{}},
		{"EC2 Net Delete", "Delete", netCommand, []string{}},
		{"EC2 Net Nuke", "Nuke", netCommand, []string{common.TestAcc, zoneID, domainID}},
		{"EC2 VPC Nuke", "Nuke", vpcCommand, []string{common.TestAcc, zoneID, domainID}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			req, err := cc.MakePlainClient(getEndpoint(port))
			assert.NoError(t, err)

			req.Cmd.Action = d.action
			req.Cmd.Command = d.command
			req.Cmd.Params = d.params

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Read" || d.action == "Delete" {
				req.Cmd.Params = append(req.Cmd.Params, netID)
			}

			if d.action == "Create" {
				req.Cmd.Params = append(req.Cmd.Params, vpcID)
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
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}

func Test_tmpl(t *testing.T) {
	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

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
		{"EC2 Tmpl List", "List", tmplCommand, []string{zoneID, domainID, common.TestAcc, common.TmplFilter}},
		{"EC2 Tmpl Create", "Create", tmplCommand, []string{common.TmplName, zoneID, domainID, common.TestAcc}},
		{"EC2 Tmpl Resolve", "Resolve", tmplCommand, []string{zoneID, domainID, common.TestAcc, common.TmplFilter, common.TmplName}},
		{"EC2 Tmpl Read", "Read", tmplCommand, []string{}},
		{"EC2 Tmpl Delete", "Delete", tmplCommand, []string{}},
		{"EC2 Tmpl Nuke", "Nuke", tmplCommand, []string{common.TestAcc, common.TmplFilter, zoneID, domainID}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			req, err := cc.MakePlainClient(getEndpoint(port))
			assert.NoError(t, err)

			req.Cmd.Action = d.action
			req.Cmd.Command = d.command
			req.Cmd.Params = d.params

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Create" {
				req.Cmd.Params = append(req.Cmd.Params, osOfferID, common.TmplURL)
			}

			if d.action == "Read" {
				req.Cmd.Params = append(req.Cmd.Params, tmplID, common.TmplFilter)
			}

			if d.action == "Delete" {
				req.Cmd.Params = append(req.Cmd.Params, tmplID, zoneID)
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
					tmplID, err = jsonparser.GetString(data, "id", "id")
					assert.NoError(t, err)
				}
			}
		})
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}

func Test_inst(t *testing.T) {
	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

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
		{"EC2 Inst List", "List", instCommand, []string{zoneID, domainID, common.TestAcc}},
		{"EC2 SSH Create", "Create", sshCommand, []string{common.SshKeyName, domainID, common.TestAcc, common.Pubkey}},
		{"EC2 VPC Create", "Create", vpcCommand, []string{common.VpcName, zoneID, domainID, common.TestAcc, vpcOfferID, common.VpcCidr4, common.NetDomain}},
		{"EC2 VPC Resolve", "Resolve", vpcCommand, []string{zoneID, domainID, common.TestAcc, common.VpcName}},
		{"EC2 Net Create", "Create", netCommand, []string{common.NetName, zoneID, domainID, common.TestAcc, common.NetCidr4, common.EmptyCIDR6, netOfferID, common.NetDomain}},
		{"EC2 Net Resolve", "Resolve", netCommand, []string{zoneID, domainID, common.TestAcc, common.NetName}},
		{"EC2 Tmpl Create", "Create", tmplCommand, []string{common.TmplName, zoneID, domainID, common.TestAcc}},
		{"EC2 Tmpl Resolve", "Resolve", tmplCommand, []string{zoneID, domainID, common.TestAcc, common.TmplFilter, common.TmplName}},
		{"EC2 Inst Create", "Create", instCommand, []string{common.InstName, zoneID, domainID, common.TestAcc}},
		{"EC2 Inst Resolve", "Resolve", instCommand, []string{zoneID, domainID, common.TestAcc, common.InstName}},
		{"EC2 Inst Read", "Read", instCommand, []string{}},
		{"EC2 Inst Delete", "Delete", instCommand, []string{}},
		{"EC2 Inst Nuke", "Nuke", instCommand, []string{common.TestAcc, zoneID, domainID}},
		{"EC2 Tmpl Nuke", "Nuke", tmplCommand, []string{common.TestAcc, common.TmplFilter, zoneID, domainID}},
		{"EC2 Net Nuke", "Nuke", netCommand, []string{common.TestAcc, zoneID, domainID}},
		{"EC2 VPC Nuke", "Nuke", vpcCommand, []string{common.TestAcc, zoneID, domainID}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			req, err := cc.MakePlainClient(getEndpoint(port))
			assert.NoError(t, err)

			req.Cmd.Action = d.action
			req.Cmd.Command = d.command
			req.Cmd.Params = d.params

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Create" {

				if d.command == netCommand {
					req.Cmd.Params = append(req.Cmd.Params, vpcID)
				}

				if d.command == tmplCommand {
					req.Cmd.Params = append(req.Cmd.Params, osOfferID, common.TmplURL)
				}

				if d.command == instCommand {
					req.Cmd.Params = append(req.Cmd.Params, tmplID, instOfferID, common.SshKeyName, fmt.Sprint(common.DiskSizeGB), fmt.Sprintf("net::%s", netID))
				}
			}

			if d.action == "Resolve" {
				if d.command == instCommand {
					req.Cmd.Params = append(req.Cmd.Params, tmplID, common.InstName)
				}
			}

			if d.action == "Read" {
				if d.command == tmplCommand {
					req.Cmd.Params = append(req.Cmd.Params, tmplID, common.TmplFilter)
				}
				if d.command == instCommand {
					req.Cmd.Params = append(req.Cmd.Params, instID)
				}
			}

			if d.action == "Delete" {
				req.Cmd.Params = append(req.Cmd.Params, instID)
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
					tmplID, err = jsonparser.GetString(data, "id", "id")
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
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}

func Test_offerings(t *testing.T) {
	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

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
		{"EC2 Zone ID", "Resolve", zoneCommand, []string{common.TestZone}},
		{"EC2 Domain ID", "Resolve", domCommand, []string{common.TestDomain}},
		{"EC2 VPC Offer ID", "Resolve", vpcOfferCommand, []string{common.VpcOffer}},
		{"EC2 Net Offer ID", "Resolve", netOfferCommand, []string{common.NetOffer}},
		{"EC2 OS Offer ID", "Resolve", osOfferCommand, []string{common.OsOffer}},
		{"EC2 Inst Offer ID", "Resolve", instOfferCommand, []string{common.InstOffer}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			req, err := cc.MakePlainClient(getEndpoint(port))
			assert.NoError(t, err)

			req.Cmd.Action = d.action
			req.Cmd.Command = d.command
			req.Cmd.Params = d.params

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
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}

func Test_acl(t *testing.T) {
	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	var zoneID, domainID, vpcOfferID, vpcID, aclID string

	tmp, err := server.kv.Get("zoneID")
	assert.NoError(t, err)

	zoneID = string(tmp.Value())

	tmp, err = server.kv.Get("domainID")
	assert.NoError(t, err)

	domainID = string(tmp.Value())

	tmp, err = server.kv.Get("vpcOfferID")
	assert.NoError(t, err)

	vpcOfferID = string(tmp.Value())

	fmt.Println(vpcID)
	fmt.Println(vpcOfferID)
	fmt.Println(aclID)

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 VPC Create", "Create", vpcCommand, []string{common.VpcName, zoneID, domainID, common.TestAcc, vpcOfferID, common.VpcCidr4, common.NetDomain}},
		{"EC2 VPC ID Resolve", "Resolve", vpcCommand, []string{zoneID, domainID, common.TestAcc, common.VpcName}},
		{"EC2 ACL Create", "Create", aclCommand, []string{common.AclName, common.AclDescr}},
		{"EC2 ACL	Resolve", "Resolve", aclCommand, []string{domainID, common.TestAcc}},
		{"EC2 ACL List", "List", aclCommand, []string{domainID, common.TestAcc}},
		{"EC2 ACL Read", "Read", aclCommand, []string{}},
		{"EC2 ACL Delete", "Delete", aclCommand, []string{}},
		{"EC2 ACL Nuke", "Nuke", aclCommand, []string{common.TestAcc, domainID}},
		{"EC2 VPC Nuke", "Nuke", vpcCommand, []string{common.TestAcc, zoneID, domainID}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			req, err := cc.MakePlainClient(getEndpoint(port))
			assert.NoError(t, err)

			req.Cmd.Action = d.action
			req.Cmd.Command = d.command
			req.Cmd.Params = d.params

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Create" {
				if d.command == aclCommand {
					req.Cmd.Params = append(req.Cmd.Params, vpcID)
				}
			}

			if d.action == "Resolve" {
				if d.command == aclCommand {
					req.Cmd.Params = append(req.Cmd.Params, vpcID, common.AclName)
				}
			}

			if d.action == "List" || d.action == "Nuke" {
				if d.command == aclCommand {
					req.Cmd.Params = append(req.Cmd.Params, vpcID)
				}
			}

			if d.action == "Read" || d.action == "Delete" {
				if d.command == aclCommand {
					req.Cmd.Params = append(req.Cmd.Params, aclID)
				}
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			if d.action == "Resolve" {

				// if d.command == osOfferCommand {
				// 	osOfferID, err = jsonparser.GetString(data, "id")
				// 	assert.NoError(t, err)
				// }

				if d.command == vpcCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

						if string(value) != jsonparser.Null.String() {
							vpcID, err = jsonparser.GetString(value, "id", "id")
							assert.NoError(t, err)
						}

						assert.NoError(t, err)
					}, "items")

					assert.NoError(t, err)
				}

				if d.command == aclCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

						if string(value) != jsonparser.Null.String() {
							aclID, err = jsonparser.GetString(value, "id", "id")
							assert.NoError(t, err)
						}

						assert.NoError(t, err)
					}, "items")

					assert.NoError(t, err)
				}
			}
		})
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}

func Test_aclrule(t *testing.T) {
	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

	// Launch Server
	server, err := MakeAPIServerMock()
	assert.NoErrorf(t, err, "Failed to create a server")

	go runServer(server.echo, port)

	var zoneID, domainID, vpcOfferID, netOfferID, vpcID, netID, aclID string

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

	fmt.Println(vpcID)
	fmt.Println(netID)
	fmt.Println(vpcOfferID)
	fmt.Println(aclID)

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"EC2 VPC Create", "Create", vpcCommand, []string{common.VpcName, zoneID, domainID, common.TestAcc, vpcOfferID, common.VpcCidr4, common.NetDomain}},
		{"EC2 VPC ID Resolve", "Resolve", vpcCommand, []string{zoneID, domainID, common.TestAcc, common.VpcName}},
		{"EC2 Net Create", "Create", netCommand, []string{common.NetName, zoneID, domainID, common.TestAcc, common.NetCidr4, common.EmptyCIDR6, netOfferID, common.NetDomain}},
		{"EC2 Net Resolve", "Resolve", netCommand, []string{zoneID, domainID, common.TestAcc, common.NetName}},
		{"EC2 ACL Create", "Create", aclCommand, []string{common.AclName, common.AclDescr}},
		{"EC2 ACL	Resolve", "Resolve", aclCommand, []string{domainID, common.TestAcc}},
		{"EC2 ACL	Rule", "Create", aclrCommand, []string{common.AclrDesc, common.AclrAction, common.AclrProto, common.AclrTraffic, common.AclrCIDR4, fmt.Sprint(common.PortStart), fmt.Sprint(common.PortEnd)}},
		{"EC2 ACL Rule List", "List", aclrCommand, []string{domainID, common.TestAcc}},
		{"EC2 ACL Nuke", "Nuke", aclCommand, []string{common.TestAcc, domainID}},
		{"EC2 Net Nuke", "Nuke", netCommand, []string{common.TestAcc, zoneID, domainID}},
		{"EC2 VPC Nuke", "Nuke", vpcCommand, []string{common.TestAcc, zoneID, domainID}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			req, err := cc.MakePlainClient(getEndpoint(port))
			assert.NoError(t, err)

			req.Cmd.Action = d.action
			req.Cmd.Command = d.command
			req.Cmd.Params = d.params

			// Append RUN Time params, which are NOT available in compile time
			if d.action == "Create" {

				if d.command == netCommand {
					req.Cmd.Params = append(req.Cmd.Params, vpcID)
				}

				if d.command == aclCommand {
					req.Cmd.Params = append(req.Cmd.Params, vpcID)
				}

				if d.command == aclrCommand {
					req.Cmd.Params = append(req.Cmd.Params, aclID, netID)
				}
			}

			if d.action == "Resolve" {
				if d.command == aclCommand {
					req.Cmd.Params = append(req.Cmd.Params, vpcID, common.AclName)
				}
			}

			if d.action == "List" || d.action == "Nuke" {
				if d.command == aclCommand {
					req.Cmd.Params = append(req.Cmd.Params, vpcID)
				}

				if d.command == aclrCommand {
					req.Cmd.Params = append(req.Cmd.Params, aclID)
				}
			}

			if d.action == "Read" || d.action == "Delete" {
				if d.command == aclCommand {
					req.Cmd.Params = append(req.Cmd.Params, aclID)
				}
			}

			res, err := req.MakeRequest()
			assert.NoErrorf(t, err, fmt.Sprintf("failed on CC client request: %v \n", err))
			assert.Equal(t, http.StatusCreated, res.StatusCode())
			assert.NotEmpty(t, res.Body)

			data, err := utils.DecodeJSONBytes(res.Body)
			assert.NoError(t, err)

			if d.action == "Resolve" {

				if d.command == vpcCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

						if string(value) != jsonparser.Null.String() {
							vpcID, err = jsonparser.GetString(value, "id", "id")
							assert.NoError(t, err)
						}

						assert.NoError(t, err)
					}, "items")

					assert.NoError(t, err)
				}

				if d.command == netCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

						if string(value) != jsonparser.Null.String() {
							netID, err = jsonparser.GetString(value, "id", "id")
							assert.NoError(t, err)
						}

						assert.NoError(t, err)
					}, "items")

					assert.NoError(t, err)
				}

				if d.command == aclCommand {

					_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

						if string(value) != jsonparser.Null.String() {
							aclID, err = jsonparser.GetString(value, "id", "id")
							assert.NoError(t, err)
						}

						assert.NoError(t, err)
					}, "items")

					assert.NoError(t, err)
				}
			}
		})
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}
