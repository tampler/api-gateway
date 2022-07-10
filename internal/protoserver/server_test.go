package protoserver

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/buger/jsonparser"
	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/internal/common"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"github.com/neurodyne-web-services/nws-sdk-go/services/natstool"
)

func Test_ssh(t *testing.T) {

	// Connect to NATS
	nc, err := natstool.MakeNatsConnect()
	assert.NoError(t, err)

	stor, err := MakeStorageServer(nc)
	assert.NoError(t, err)

	// Build a global config
	var cfg config.AppConfig

	if err := cfg.AppInit(CONFIG_NAME, CONFIG_PATH); err != nil {
		log.Fatalf("Config failed %s", err.Error())
	}

	logger, _ := logging.MakeLogger(cfg.Log.Verbosity, cfg.Log.Output)
	defer logger.Sync()
	zl := logger.Sugar()

	app, err := buildProtoServer(nc, cfg, zl)
	assert.NoError(t, err)

	tmp, err := stor.kv.Get("domainID")
	assert.NoError(t, err)

	domainID := string(tmp.Value())

	var sshKeyID string

	data := []struct {
		name   string
		action string
		params []string
	}{
		{"EC2 SSH", "List", []string{domainID, common.TestAcc}},
		{"EC2 SSH", "Create", []string{common.SshKeyName, domainID, common.TestAcc, common.Pubkey}},
		{"EC2 SSH", "Resolve", []string{domainID, common.TestAcc, common.SshKeyName}},
		{"EC2 SSH", "Read", []string{}},
		{"EC2 SSH", "Delete", []string{common.SshKeyName, domainID, common.TestAcc}},
		{"EC2 SSH", "Nuke", []string{common.TestAcc, domainID}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			cmd := cc.APICommand{
				Service:  "EC2",
				Resource: "SSHKeypair",
				Action:   d.action,
				Params:   d.params,
			}

			if d.action == "Read" || d.action == "Delete" {
				cmd.Params = append(cmd.Params, sshKeyID)
			}

			req := cc.APIRequest{
				JobID: uuid.NewString(),
				Cmd:   &cmd,
			}

			res, err := app.UnaryCall(context.Background(), &req)
			assert.NoError(t, err)

			// Resolve runtime VPC ID
			if d.action == "Resolve" {
				sshKeyID, err = jsonparser.GetString(res.Data, "id", "id")
				assert.NoError(t, err)
			}
		})
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}

func Test_failer(t *testing.T) {

	// Connect to NATS
	nc, err := natstool.MakeNatsConnect()
	assert.NoError(t, err)

	// Build a global config
	var cfg config.AppConfig

	if err := cfg.AppInit(CONFIG_NAME, CONFIG_PATH); err != nil {
		log.Fatalf("Config failed %s", err.Error())
	}

	logger, _ := logging.MakeLogger(cfg.Log.Verbosity, cfg.Log.Output)
	defer logger.Sync()
	zl := logger.Sugar()

	app, err := buildProtoServer(nc, cfg, zl)
	assert.NoError(t, err)

	data := []struct {
		name   string
		action string
		params []string
	}{
		{"Failer BlackHole", "List", []string{}},
		// {"EC2 SSH Create", "Create", []string{}},
		// {"EC2 SSH Resolve", "Resolve", []string{}},
		// {"EC2 SSH Read", "Read", []string{}},
		// {"EC2 SSH Delete", "Delete", []string{}},
		// {"EC2 SSH Nuke", "Nuke", []string{}},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {

			cmd := cc.APICommand{
				Service:  "Failer",
				Resource: "BlackHole",
				Action:   d.action,
				Params:   d.params,
			}

			req := cc.APIRequest{
				JobID: uuid.NewString(),
				Cmd:   &cmd,
			}

			_, err := app.UnaryCall(context.Background(), &req)
			assert.NoError(t, err)

		})
		time.Sleep(common.SleepTime * time.Millisecond)
	}
}
