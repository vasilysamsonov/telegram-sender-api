package middleware

import (
	"github.com/gofiber/fiber/v2"
	fiberrequestid "github.com/gofiber/fiber/v2/middleware/requestid"
)

func RequestID() fiber.Handler {
	return fiberrequestid.New()
}

func GetRequestID(c *fiber.Ctx) string {
	requestID := c.GetRespHeader(fiber.HeaderXRequestID)
	if requestID != "" {
		return requestID
	}

	requestID = c.Get(fiber.HeaderXRequestID)
	if requestID != "" {
		return requestID
	}

	if local := c.Locals("requestid"); local != nil {
		if value, ok := local.(string); ok {
			return value
		}
	}

	return ""
}
