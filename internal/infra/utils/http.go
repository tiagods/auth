package utils

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"net/http"
)

func JSON(e echo.Context, code int, err error) error {
	resultCode := code
	message := http.StatusText(http.StatusInternalServerError)
	resultErr := err

	var httpError httperrors.HttpError
	if ok := errors.As(err, &httpError); ok {
		message = httpError.Error()
		resultCode = httpError.StatusCode
		resultErr = httpError.InternalError
	}

	if resultCode >= http.StatusInternalServerError {
		e.Logger().Error(resultErr)
	} else {
		e.Logger().Warn(resultErr)
	}

	return e.JSON(code, echo.HTTPError{Code: resultCode, Message: message, Internal: resultErr})
}
