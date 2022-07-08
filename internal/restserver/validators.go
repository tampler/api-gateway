package restserver

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/neurodyne-web-services/api-gateway/internal/token"
)

const (
	delimiter = "::"
)

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(cmd interface{}) error {
	// Validate Command format
	if err := token.CommandValidator(cmd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}
