package handler

import (
	"fmt"
	"net/http"

	"github.com/tiagods/auth/internal/adapter/web/extractor"
	"github.com/tiagods/auth/internal/infra/logger"
	"github.com/tiagods/auth/internal/infra/tracer"

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
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	register := &request.Register{}
	if err := extractor.Extractor(c, register); err != nil {
		tracer.SetError(span, err)
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Request POST at %s", c.Path()))

	result, err := h.service.Register(ctx, register)
	if err != nil {
		tracer.SetError(span, err)
		return err
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *TokenHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	caller := "handler::login"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	login := &request.Login{}

	if err := extractor.Extractor(c, login); err != nil {
		tracer.SetError(span, err)
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Request POST at %s", c.Path()))

	result, err := h.service.Login(ctx, login)
	if err != nil {
		tracer.SetError(span, err)
		return err
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *TokenHandler) RevokeToken(c echo.Context) error {
	ctx := c.Request().Context()
	caller := "handler::revoke_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	tokenReq := &request.RefreshToken{}
	if err := extractor.Extractor(c, tokenReq); err != nil {
		tracer.SetError(span, err)
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Request DELETE at %s", c.Path()))

	if err := h.service.RevokeToken(ctx, tokenReq); err != nil {
		tracer.SetError(span, err)
		return err
	}
	return c.NoContent(http.StatusOK)
}

func (h *TokenHandler) RecreateToken(c echo.Context) error {
	ctx := c.Request().Context()
	caller := "handler::refresh_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	tokenReq := &request.RefreshToken{}
	if err := extractor.Extractor(c, tokenReq); err != nil {
		tracer.SetError(span, err)
		return err
	}

	logger.Info(ctx, fmt.Sprintf("Request POST at %s", c.Path()))

	result, err := h.service.RecreateToken(ctx, tokenReq)
	if err != nil {
		tracer.SetError(span, err)
		return err
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *TokenHandler) ValidateToken(c echo.Context) error {
	ctx := c.Request().Context()
	caller := "handler::validate_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	authorization := c.Request().Header.Get("Authorization")

	logger.Info(ctx, fmt.Sprintf("Request GET at %s", c.Path()))

	if err := h.service.ValidateToken(ctx, authorization); err != nil {
		tracer.SetError(span, err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "token is valid"})
}
