package handler

import (
	"github.com/tiagods/auth/internal/adapter/web/presenter/response"
	"github.com/tiagods/auth/internal/domain/service"
	"github.com/tiagods/auth/internal/infra/otel"
	"go.opentelemetry.io/otel/codes"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	health service.Health
}

func NewHealthHandler(health service.Health) HealthHandler {
	return HealthHandler{health: health}
}

func (h *HealthHandler) Health(c echo.Context) error {
	ctx, span := otel.Start(c.Request().Context(), "handler::health", otel.SpanKindRequest)
	defer span.End()

	result := h.health.Check(ctx)
	responseResult := response.FromEntity(result)
	statusCode := http.StatusInternalServerError
	if responseResult.Status {
		span.SetStatus(codes.Ok, "health check ok")
		statusCode = http.StatusOK
	}
	return c.JSON(statusCode, responseResult)
}
