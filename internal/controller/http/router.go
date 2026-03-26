package http

import (
	"net/http"

	"telegram-sender-api/internal/controller/http/middleware"
	v1 "telegram-sender-api/internal/controller/http/v1"
	"telegram-sender-api/internal/usecase"
	"telegram-sender-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

func NewRouter(app *fiber.App, m usecase.Message, l logger.Interface) {
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(l))
	app.Use(middleware.Recovery(l))

	app.Get("/healthz", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(http.StatusOK)
	})

	apiV1Group := app.Group("/v1")
	{
		v1.NewMessageRoutes(apiV1Group, m, l)
	}
}
