package middleware

import (
	"fmt"

	"telegram-sender-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
)

func Recovery(l logger.Interface) fiber.Handler {
	return fiberrecover.New(fiberrecover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, recovered interface{}) {
			l.Error(fmt.Sprintf("request_id=%s panic: %v", GetRequestID(c), recovered))
		},
	})
}
