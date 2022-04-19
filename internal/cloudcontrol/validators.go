package cloudcontrol

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/nws-sdk-go/pkg/utils"
)

const (
	delimiter              = "::"
	defaultSliceLen        = 4
	defaultCommandCapacity = 8
)

type CustomValidator struct {
	Validator *validator.Validate
}

func availableServices() []string {
	out := make([]string, 2, defaultCommandCapacity)

	out = append(out, "S3")
	out = append(out, "EC2")

	return out
}

// availableResources - list resources from all available services
func availableResources() []string {
	out := make([]string, 2, defaultSliceLen)

	// out = append(out, s3.AvailableResources()...)
	// out = append(out, ec2.AvailableResources()...)

	return out
}

func commandValidator(command interface{}) error {
	cmd := strings.Split(command.(string), delimiter)

	if cmd[0] != "NWS" {
		return fmt.Errorf("invalid provider specified. Default: NWS")
	}

	if !utils.Contains(availableServices(), cmd[1]) {
		return fmt.Errorf("invalid service name specified")
	}

	if !utils.Contains(availableResources(), cmd[2]) {
		return fmt.Errorf("invalid resource name specified")
	}

	return nil
}

func (cv *CustomValidator) Validate(cmd interface{}) error {
	// Validate Command format
	if err := commandValidator(cmd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}
