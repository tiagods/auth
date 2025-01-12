package logger

import (
	"context"
	"github.com/tiagods/auth/internal/infra/utils"
	"go.uber.org/zap"
)

type (
	Logger struct {
	}

	Fields struct {
		key   string
		value any
	}
)

func Init() *zap.Logger {
	zap.ReplaceGlobals(zap.Must(zap.NewProductionConfig().Level.SetLevel(zap.DebugLevel)))
	zap.L().Info("logger construction succeeded")
	return zap.L()
}

func Error(ctx context.Context, err error, message string, fields ...Fields) {
	zap.L().Error(message, mapToFields(ctx, err, fields)...)
}

func Debug(ctx context.Context, message string, fields ...Fields) {
	zap.L().Debug(message, mapToFields(ctx, nil, fields)...)
}

func Info(ctx context.Context, message string, fields ...Fields) {
	zap.L().Info(message, mapToFields(ctx, nil, fields)...)
}

func Core(ctx context.Context, message string, fields ...Fields) {
	zap.L().Info(message, mapToFields(ctx, nil, fields)...)
}

func Warn(ctx context.Context, err error, message string, fields ...Fields) {
	zap.L().Warn(message, mapToFields(ctx, err, fields)...)
}

func Fatal(ctx context.Context, err error, message string, fields ...Fields) {
	zap.L().Fatal(message, mapToFields(ctx, err, fields)...)
}

func mapToFields(ctx context.Context, err error, fields []Fields) []zap.Field {
	var result []zap.Field
	for _, v := range fields {
		result = append(result, zap.Any(v.key, v.value))
	}

	result = append(result, zap.String("cid", utils.GetCidFromContext(ctx)))
	result = append(result, zap.String("orgId", utils.GetTenantFromContext(ctx)))

	if err != nil {
		result = append(result, zap.Error(err))
		result = append(result, zap.String("errorMessage", err.Error()))
	}
	return result
}
