package middleware

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"github.com/tiagods/auth/internal/infra/requestcontext"
	"net/http"
)

func RequestContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cid := c.Request().Header.Get(requestcontext.CID)
		if cid == "" {
			cid = uuid.NewString()
		}
		tenant := c.Request().Header.Get(requestcontext.Tenant)

		c.Response().Header().Set(echo.HeaderServer, "Echo/3.0")
		ctx := c.Request().Context()
		ctx = context.WithValue(ctx, requestcontext.ContextKey, requestcontext.RequestContext{
			Cid:    cid,
			Tenant: tenant,
			Roles:  nil,
		})
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	resultCode := http.StatusInternalServerError
	message := http.StatusText(http.StatusInternalServerError)
	resultErr := err

	var he *echo.HTTPError
	if errors.As(err, &he) {
		resultCode = he.Code
	}
	var httpError *httperrors.HttpError
	if ok := errors.As(err, httpError); ok {
		message = httpError.Error()
		resultCode = httpError.StatusCode
		resultErr = httpError.InternalError
	}

	if resultCode >= http.StatusInternalServerError {
		c.Logger().Error(resultErr)
	} else {
		c.Logger().Warn(resultErr)
	}

	hte := echo.HTTPError{Code: resultCode, Message: message, Internal: resultErr}
	_ = c.JSON(hte.Code, hte)
}
