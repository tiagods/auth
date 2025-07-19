package httperrors

import (
	"context"
	"net/http"

	"github.com/tiagods/auth/internal/infra/logger"
	"github.com/tiagods/auth/internal/infra/message"
)

type HttpError struct {
	StatusCode    int                  `json:"status_code"`
	Code          string               `json:"code"`
	Message       message.ErrorMessage `json:"message"`
	InternalError error
}

func (h HttpError) Error() string {
	return h.Code + ":" + h.Message.UserMessage
}

func NewHttpError(ctx context.Context, statusCode int, message message.ErrorMessage, err error, fields ...logger.Fields) HttpError {
	if err == nil {
		err = message.GetError()
	}

	if statusCode >= http.StatusInternalServerError {
		logger.Error(ctx, err, message.UserMessage, fields...)
	} else {
		logger.Warn(ctx, err, message.UserMessage, fields...)
	}

	return HttpError{
		StatusCode:    statusCode,
		Code:          message.Code,
		Message:       message,
		InternalError: err,
	}
}
