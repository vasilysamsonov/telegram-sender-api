package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"telegram-sender-api/internal/controller/http/middleware"
	"telegram-sender-api/internal/controller/http/v1/request"
	"telegram-sender-api/internal/controller/http/v1/response"
	"telegram-sender-api/internal/entity"
	"telegram-sender-api/internal/usecase"
	"telegram-sender-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type V1 struct {
	m usecase.Message
	l logger.Interface
}

const maxRequestBodyBytes = 1 << 20

func (r *V1) sendMessage(c *fiber.Ctx) error {
	if len(c.Body()) > maxRequestBodyBytes {
		r.logRequestError(c, fiber.StatusRequestEntityTooLarge, "request_too_large", fmt.Errorf("request body exceeds %d bytes", maxRequestBodyBytes))
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(response.SendMessage{
			Status: "error",
			Error:  "request body is too large",
		})
	}

	var req request.SendMessage
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		r.logRequestError(c, fiber.StatusBadRequest, transportErrorType(err), err)
		return c.Status(fiber.StatusBadRequest).JSON(response.SendMessage{
			Status: "error",
			Error:  "invalid json body",
		})
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		r.logRequestError(c, fiber.StatusBadRequest, "multiple_json_objects", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.SendMessage{
			Status: "error",
			Error:  "request body must contain a single JSON object",
		})
	}

	err := r.m.Send(c.UserContext(), req.BotToken, entity.Message{
		ChatID: int64(req.ChatID),
		Text:   req.Text,
	})
	if err != nil {
		status := httpStatus(err)
		r.logRequestError(c, status, errorType(err), err)
		return c.Status(status).JSON(response.SendMessage{
			Status: "error",
			Error:  clientErrorMessage(err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.SendMessage{
		Status: "ok",
	})
}

func (r *V1) logRequestError(c *fiber.Ctx, status int, kind string, err error) {
	r.l.Error(fmt.Errorf(
		"request_id=%s method=%s path=%s status=%d error_type=%s error=%q payload=%s",
		middleware.GetRequestID(c),
		c.Method(),
		c.Path(),
		status,
		kind,
		err.Error(),
		middleware.LoggedPayload(c),
	))
}
