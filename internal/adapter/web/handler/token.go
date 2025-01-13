package handler

import (
	"fmt"
	"github.com/tiagods/auth/internal/adapter/web/extractor"
	"github.com/tiagods/auth/internal/infra/logger"
	"github.com/tiagods/auth/internal/infra/otel"
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
	ctx := c.Request().Context()
	caller := "handler::register"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()
	return nil
}

func (h *TokenHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	caller := "handler::login"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	login := &request.Login{}

	if err := extractor.ExtractPresenter(c, login); err != nil {
		otel.SetError(span, err)
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Request POST at %s", c.Path()))

	result, err := h.service.Login(ctx, login)
	if err != nil {
		otel.SetError(span, err)
		return err
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *TokenHandler) RefreshToken(c echo.Context) error {
	ctx := c.Request().Context()
	caller := "handler::refresh_token"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	tokenReq := &request.RefreshToken{}
	if err := extractor.ExtractPresenter(c, tokenReq); err != nil {
		otel.SetError(span, err)
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Request POST at %s", c.Path()))

	result, err := h.service.RefreshToken(ctx, tokenReq)
	if err != nil {
		otel.SetError(span, err)
		return err
	}
	return c.JSON(http.StatusCreated, result)
}
