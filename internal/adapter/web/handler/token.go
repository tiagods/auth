package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tiagods/auth/internal/adapter/web/presenter/request"
	"github.com/tiagods/auth/internal/domain/service"
	"github.com/tiagods/auth/internal/infra/utils"
)

type (
	TokenHander struct {
		service service.TokenService
	}
)

func NewTokenHandler(service service.TokenService) *TokenHander {
	return &TokenHander{service: service}
}

func (h TokenHander) Register(c echo.Context) error {
	return nil
}

func (h TokenHander) Login(c echo.Context) error {
	login := &request.Login{}

	if err := extractPresenter(c, login); err != nil {
		return err
	}

	result, err := h.service.Login(c.Request().Context(), login)
	if err != nil {
		return utils.JSON(c, 0, err)
	}

	return c.JSON(http.StatusCreated, result)
}

func extractPresenter(c echo.Context, i interface{}) error {
	if err := c.Bind(i); err != nil {
		c.Logger().Error(err)
		return utils.JSON(c, http.StatusBadRequest, err)
	}
	if err := c.Validate(i); err != nil {
		c.Logger().Error(err)
		return utils.JSON(c, http.StatusBadRequest, err)
	}
	return nil
}

func (h TokenHander) RefreshToken(c echo.Context) error {
	tokenReq := &request.RefreshToken{}
	if err := extractPresenter(c, tokenReq); err != nil {
		return err
	}

	result, err := h.service.RefreshToken(c.Request().Context(), tokenReq)
	if err != nil {
		return utils.JSON(c, 0, err)
	}
	return c.JSON(http.StatusCreated, result)
}
