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

func Test_sess(t *testing.T) {
	t.Skip()
	// os.Setenv("NATS_URL", "192.168.1.93:42377")
	// os.Setenv("NATS_USER", "local")
	// os.Setenv("NATS_PASS", "vsO2TcFwkmQ2p3eiOl3HcD7NVWEjvRI4")

	port := rand.Intn(common.PortEnd-common.PortStart) + common.PortStart

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

	now := time.Now().Format(common.TimeFormat)

	data := []struct {
		name    string
		action  string
		command string
		params  []string
	}{
		{"Session Create", "Create", sessCommand, []string{common.TestZone, zoneID, common.TestDomain, domainID, account, common.NetDomain, now, userID}},
		{"Session Read", "Read", sessCommand, []string{userID}},
		{"Session Delete", "Delete", sessCommand, []string{userID}},
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

			fmt.Println(string(data))

			if req.Cmd.Action == "Read" {

				readZone, err := jsonparser.GetString(data, "zone", "name")
				assert.NoError(t, err)
				assert.Equal(t, common.TestZone, readZone)

				readZoneID, err := jsonparser.GetString(data, "zone", "id")
				assert.NoError(t, err)
				assert.Equal(t, zoneID, readZoneID)

				readDomain, err := jsonparser.GetString(data, "domain", "name")
				assert.NoError(t, err)
				assert.Equal(t, common.TestDomain, readDomain)

				readDomainID, err := jsonparser.GetString(data, "domain", "id")
				assert.NoError(t, err)
				assert.Equal(t, domainID, readDomainID)

				readAccount, err := jsonparser.GetString(data, "account")
				assert.NoError(t, err)
				assert.Equal(t, account, readAccount)
			}
		})
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}
