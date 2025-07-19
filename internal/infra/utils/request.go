package utils

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tiagods/auth/internal/infra/requestcontext"
)

func GetCidFromContext(ctx context.Context) string {
	cid := ""
	if rq, ok := ctx.Value(requestcontext.ContextKey).(requestcontext.RequestContext); ok {
		if rq.Cid == "" {
			rq.Cid = uuid.NewString()
			context.WithValue(ctx, requestcontext.ContextKey, rq)
		}
		cid = rq.Cid
	}
	return cid
}

func GetTenantFromContext(ctx context.Context) string {
	tenant := ""
	if rq, ok := ctx.Value(requestcontext.ContextKey).(requestcontext.RequestContext); ok {
		tenant = rq.Tenant
	}
	return tenant
}

func GetCidFromEchoContext(eCtx echo.Context) string {
	return GetCidFromContext(eCtx.Request().Context())
}

func GetTenantFromEchoContext(eCtx echo.Context) string {
	return GetTenantFromContext(eCtx.Request().Context())
}

func GetRequestContext(eCtx echo.Context) requestcontext.RequestContext {
	if rq, ok := eCtx.Request().Context().Value(requestcontext.ContextKey).(requestcontext.RequestContext); ok {
		return rq
	}
	return requestcontext.RequestContext{}
}
