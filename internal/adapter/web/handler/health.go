package handler

import (
	"net/http"

	"github.com/tiagods/auth/internal/adapter/web/presenter/response"
	"github.com/tiagods/auth/internal/domain/service"
	"github.com/tiagods/auth/internal/infra/tracer"
	"go.opentelemetry.io/otel/codes"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	health service.Health
}

func NewHealthHandler(health service.Health) HealthHandler {
	return HealthHandler{health: health}
}

func (h *HealthHandler) Health(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "handler::health", tracer.SpanKindRequest)
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
