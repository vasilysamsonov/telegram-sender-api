package v1

import (
	"errors"
	"strings"

	"telegram-sender-api/internal/usecase/message"

	"github.com/gofiber/fiber/v2"
)

func httpStatus(err error) int {
	switch {
	case errors.Is(err, message.ErrInvalidInput):
		return fiber.StatusBadRequest
	case errors.Is(err, message.ErrExternal):
		return fiber.StatusBadGateway
	default:
		return fiber.StatusInternalServerError
	}
}

func clientErrorMessage(err error) string {
	switch {
	case errors.Is(err, message.ErrInvalidInput):
		return "invalid request data"
	case errors.Is(err, message.ErrExternal):
		return "failed to send message"
	default:
		return "internal server error"
	}
}

func errorType(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, message.ErrInvalidInput):
		return "validation_error"
	case errors.Is(err, message.ErrExternal):
		return "external_error"
	default:
		return "internal_error"
	}
}

func transportErrorType(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()

	switch {
	case strings.Contains(msg, "unknown field"):
		return "unknown_field"
	case strings.Contains(msg, "cannot unmarshal"):
		return "invalid_json_type"
	case strings.Contains(msg, "unexpected EOF"):
		return "invalid_json"
	case strings.Contains(msg, "EOF"):
		return "empty_body"
	default:
		return "invalid_json"
	}
}
