package handler

import (
	"fmt"
	"github.com/tiagods/auth/internal/adapter/web/extractor"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"github.com/tiagods/auth/internal/infra/logger"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tiagods/auth/internal/adapter/web/presenter/request"
	"github.com/tiagods/auth/internal/domain/service"
)

type (
	TokenHandler struct {
		service service.TokenService
	}
)

func NewTokenHandler(service service.TokenService) *TokenHandler {
	return &TokenHandler{service: service}
}

func (h *TokenHandler) Register(c echo.Context) error {
	return nil
}

func (h *TokenHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	login := &request.Login{}

	if err := extractor.ExtractPresenter(c, login); err != nil {
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Request POST at %s", c.Path()))

	result, err := h.service.Login(c.Request().Context(), login)
	if err != nil {
		return httperrors.JSON(c, 0, err)
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *TokenHandler) RefreshToken(c echo.Context) error {
	ctx := c.Request().Context()

	tokenReq := &request.RefreshToken{}
	if err := extractor.ExtractPresenter(c, tokenReq); err != nil {
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Request POST at %s", c.Path()))

	result, err := h.service.RefreshToken(c.Request().Context(), tokenReq)
	if err != nil {
		return httperrors.JSON(c, 0, err)
	}
	return c.JSON(http.StatusCreated, result)
}
