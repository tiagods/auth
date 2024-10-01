package httperrors

import "strconv"

type HttpError struct {
	StatusCode    int    `json:"status_code"`
	Code          string `json:"code"`
	Message       string `json:"message"`
	InternalError error
}

func (h HttpError) Error() string {
	return h.Code + ":" + h.Message
}

func NewHttpError(statusCode int, message string, err error) HttpError {
	return HttpError{
		StatusCode:    statusCode,
		Code:          strconv.Itoa(statusCode),
		Message:       message,
		InternalError: err,
	}
}
