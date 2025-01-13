package service

import (
	"context"
	"github.com/labstack/gommon/log"
	"github.com/tiagods/auth/internal/adapter/database"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/cache"
	"github.com/tiagods/auth/internal/infra/otel"
	"github.com/tiagods/auth/internal/infra/requestcontext"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Health struct {
	repo  database.Repository
	cache cache.Repository
}

const (
	Database = "database"
	Cache    = "cache"
)

func NewHealthService(repo database.Repository, cache cache.Repository) Health {
	return Health{
		repo:  repo,
		cache: cache,
	}
}

func (h Health) Check(ctx context.Context) entity.Health {
	caller := "service::health"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	health := entity.NewHealth()
	health.Status = true
	health.Service[Database] = true
	health.Service[Cache] = h.cache.IsAlive()

	span.SetAttributes(attribute.String("service", "health"))

	if _, ok := ctx.Value(requestcontext.ContextKey).(requestcontext.RequestContext); ok {
		log.Info("context key is ok")
	} else {
		log.Info("context key is not ok")
	}

	err := h.repo.Ping()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		health.Service[Database] = false
		log.Error(err)
	}

	for _, service := range health.Service {
		if !service {
			health.Status = false
			break
		}
	}

	span.AddEvent(caller,
		trace.WithAttributes(
			attribute.Bool("status", health.Status),
		))
	return health
}
