package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	presenter "github.com/tiagods/auth/internal/adapter/web/presenter/response"
)

func Health(c echo.Context) error {
	h := presenter.Health{Status: "ok"}
	return c.JSON(http.StatusOK, h)
}
