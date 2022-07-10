package protoserver

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/buger/jsonparser"
	aj "github.com/choria-io/asyncjobs"
	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/internal/common"
	"github.com/neurodyne-web-services/api-gateway/internal/config"
	"github.com/neurodyne-web-services/api-gateway/internal/logging"
	"github.com/neurodyne-web-services/api-gateway/internal/worker"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"github.com/neurodyne-web-services/nws-sdk-go/services/natstool"
)

const (
	CONFIG_PATH = "../../configs"
	CONFIG_NAME = "app"
)

func Test_unary(t *testing.T) {
	// Build a global config
	var cfg config.AppConfig

	if err := cfg.AppInit(CONFIG_NAME, CONFIG_PATH); err != nil {
		log.Fatalf("Config failed %s", err.Error())
	}

	logger, _ := logging.MakeLogger(cfg.Log.Verbosity, cfg.Log.Output)
	defer logger.Sync()
	zl := logger.Sugar()

	// Connect to NATS
	nc, err := natstool.MakeNatsConnect()
	assert.NoError(t, err)

	// Input queue
	pingClient, err := aj.NewClient(
		aj.NatsConn(nc),
		aj.BindWorkQueue(cfg.Ajc.Ingress.Name),
		aj.ClientConcurrency(cfg.Ajc.Ingress.Concurrency),
		aj.PrometheusListenPort(cfg.Ajc.Ingress.MetricsPort),
		aj.RetryBackoffPolicy(aj.RetryLinearOneMinute))

	assert.NoError(t, err)

	// Output queue
	pongClient, err := aj.NewClient(
		aj.NatsConn(nc),
		aj.BindWorkQueue(cfg.Ajc.Egress.Name),
		aj.ClientConcurrency(cfg.Ajc.Egress.Concurrency),
		aj.PrometheusListenPort(cfg.Ajc.Egress.MetricsPort),
		aj.RetryBackoffPolicy(aj.RetryLinearOneMinute))

	assert.NoError(t, err)

	// Create queue routers
	pingRouter := aj.NewTaskRouter()
	pongRouter := aj.NewTaskRouter()

	pingMgr := worker.MakeQueueManager(pingClient, pingRouter)
	pongMgr := worker.MakeQueueManager(pongClient, pongRouter)

	pub := worker.MakePublisher(pongMgr, zl, map[uuid.UUID]worker.Subscriber{})
	pub.AddHandlers(cfg.Ajc.Egress.Topic)

	stor, err := MakeStorageServer(nc, cfg)
	assert.NoError(t, err)

	app := MakeProtoServer(&cfg, zl, pingMgr, pongMgr, &pub)

	tmp, err := stor.kv.Get("domainID")
	assert.NoError(t, err)

	domainID := string(tmp.Value())

	var sshKeyID string

	data := []struct {
		name   string
		action string
		params []string
	}{
		{"EC2 SSH List", "List", []string{domainID, common.TestAcc}},
		{"EC2 SSH Create", "Create", []string{common.SshKeyName, domainID, common.TestAcc, common.Pubkey}},
		{"EC2 SSH Resolve", "Resolve", []string{domainID, common.TestAcc, common.SshKeyName}},
		{"EC2 SSH Read", "Read", []string{}},
		{"EC2 SSH Delete", "Delete", []string{common.SshKeyName, domainID, common.TestAcc}},
		{"EC2 SSH Nuke", "Nuke", []string{common.TestAcc, domainID}},
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
