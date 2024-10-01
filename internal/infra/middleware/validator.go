package token_middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
)

type (
	customValidate struct {
		validator *validator.Validate
	}

	Validator interface {
		Validate(i interface{}) error
	}
)

func NewValidator() Validator {
	return &customValidate{validator: validator.New()}
}

func (cv *customValidate) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
