package token

import (
	"fmt"
	"strings"

	"github.com/neurodyne-web-services/nws-sdk-go/pkg/utils"
	"github.com/neurodyne-web-services/nws-sdk-go/services/ec2"
	"github.com/neurodyne-web-services/nws-sdk-go/services/session"
)

const (
	delimiter       = "::"
	defaultSliceLen = 8
)

func AvailableServices() []string {
	out := make([]string, defaultSliceLen)

	out = append(out, "EC2")
	out = append(out, "Session")

	return out
}

// availableResources - list resources from all available services
func AvailableResources() []string {
	out := make([]string, 2, 32)

	// out = append(out, s3.AvailableResources()...)
	out = append(out, ec2.AllResources...)
	out = append(out, session.AllResources...)

	return out
}

func CommandValidator(command string) error {
	cmd := strings.Split(command, delimiter)

	if cmd[0] != "NWS" {
		return fmt.Errorf("invalid provider specified. Default: NWS")
	}

	if !utils.Contains(AvailableServices(), cmd[1]) {
		return fmt.Errorf("invalid service name specified")
	}

	if !utils.Contains(AvailableResources(), cmd[2]) {
		return fmt.Errorf("invalid resource name specified")
	}

	return nil
}
