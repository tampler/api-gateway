package protoserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/neurodyne-web-services/api-gateway/pkg/genout/cc"
	"github.com/stretchr/testify/assert"
)

func Test_unary(t *testing.T) {
	srv := protoServer{}

	cmd := cc.APICommand{
		Service:  "EC2",
		Resource: "SSHKeypair",
		Action:   "List",
		Params:   []string{},
	}

	req := cc.APIRequest{
		Cmd:   &cmd,
		JobID: uuid.NewString(),
	}

	res, err := srv.UnaryCall(context.Background(), &req)
	assert.NoError(t, err)

	fmt.Println(res)
}
