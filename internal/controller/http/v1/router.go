package v1

import (
	"telegram-sender-api/internal/usecase"
	"telegram-sender-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

func NewMessageRoutes(apiV1Group fiber.Router, m usecase.Message, l logger.Interface) {
	r := &V1{m: m, l: l}

	messageGroup := apiV1Group.Group("/messages")
	{
		messageGroup.Post("/send", r.sendMessage)
	}
}
