package message

import (
	"errors"
	"fmt"
)

type (
	ErrorMessage struct {
		Code        string `json:"code"`
		UserMessage string `json:"message"`
	}
)

func (e ErrorMessage) GetError() error {
	return errors.New(e.Error())
}

func (e ErrorMessage) Error() string {
	return e.Code + ":" + e.UserMessage
}

func (e ErrorMessage) Format(input ...any) ErrorMessage {
	e.UserMessage = fmt.Sprintf(e.UserMessage, input...)
	return e
}

const SERVICE = "AUTH"
const ERR_PREFIX = "E" + SERVICE
const WARN_PREFIX = "W" + SERVICE
const INFO_PREFIX = "I" + SERVICE

var (
	ErrLoginRequired = ErrorMessage{
		Code:        WARN_PREFIX + "0001",
		UserMessage: "Login required",
	}

	ErrInvalidCredentials = ErrorMessage{
		Code:        WARN_PREFIX + "0002",
		UserMessage: "Invalid credentials",
	}
	ErrInvalidToken = ErrorMessage{
		Code:        WARN_PREFIX + "0002",
		UserMessage: "Invalid token",
	}
	ErrTokenExpired = ErrorMessage{
		Code:        WARN_PREFIX + "0003",
		UserMessage: "ID expired",
	}
	ErrUserNotFound = ErrorMessage{
		Code:        WARN_PREFIX + "0004",
		UserMessage: "Invalid User or password",
	}

	ErrDuplicateUser = ErrorMessage{
		Code:        ERR_PREFIX + "0005",
		UserMessage: "User already exists",
	}

	ErrRefreshNotFound = ErrorMessage{
		Code:        WARN_PREFIX + "0006",
		UserMessage: "Refresh token not found",
	}
)
