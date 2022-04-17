package cloudcontrol

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/nws-sdk-go/services/cloudcontrol/api"
)

type CCServer struct{}

// FIXME - reimplement this with a Pool
func MakeCCServer() *CCServer {
	var srv CCServer

	return &srv
}

func (c *CCServer) GetMetrics(ctx echo.Context) error {
	return sendCloudControlError(ctx, http.StatusInternalServerError, "NYI - not yet implemented")
}

func (c *CCServer) PostV1(ctx echo.Context) error {
	return sendCloudControlError(ctx, http.StatusInternalServerError, "NYI - not yet implemented")
}

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.``
func sendCloudControlError(ctx echo.Context, code int, message string) error {
	petErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, petErr)
	return err
}
