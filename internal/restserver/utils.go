package restserver

import (
	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/api-gateway/pkg/rest"
)

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.``
func sendAPIError(ctx echo.Context, code int, message string) error {
	petErr := rest.Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, petErr)
	return err
}
