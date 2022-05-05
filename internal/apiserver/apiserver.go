package apiserver

import (
	aj "github.com/choria-io/asyncjobs"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/fail"
	"go.uber.org/zap"
)

func BuildQueueManger(queueName string, zl *zap.SugaredLogger) (QueueManager, error) {
	var empty QueueManager

	client, err := aj.NewClient(
		aj.NatsContext("AJC"),
		aj.BindWorkQueue(queueName),
		aj.ClientConcurrency(10),
		// aj.CustomLogger(zl),
		// aj.PrometheusListenPort(8089),
		aj.RetryBackoffPolicy(aj.RetryLinearOneMinute))

	if err != nil {
		return empty, err
	}

	rtr := aj.NewTaskRouter()
	if rtr == nil {
		return empty, fail.Error500("Failed to create a queue router")
	}

	return MakeQueueManager(client, rtr), nil
}
