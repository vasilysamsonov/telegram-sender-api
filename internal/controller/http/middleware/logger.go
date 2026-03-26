package middleware

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"telegram-sender-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

const maxLoggedBodyBytes = 512

func Logger(l logger.Interface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		payload := LoggedPayload(c)
		l.Info(
			"request_id=%s method=%s path=%s status=%d duration=%s payload=%s",
			GetRequestID(c),
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			time.Since(start).String(),
			payload,
		)
		return err
	}
}

func LoggedPayload(c *fiber.Ctx) string {
	if len(c.Body()) == 0 {
		return "-"
	}

	if !strings.Contains(strings.ToLower(c.Get(fiber.HeaderContentType)), fiber.MIMEApplicationJSON) {
		return truncateForLog(string(c.Body()))
	}

	var payload any
	if err := json.NewDecoder(bytes.NewReader(c.Body())).Decode(&payload); err != nil {
		return truncateForLog(string(c.Body()))
	}

	maskSensitiveFields(payload)

	b, err := json.Marshal(payload)
	if err != nil {
		return "-"
	}

	return truncateForLog(string(b))
}

func maskSensitiveFields(v any) {
	switch item := v.(type) {
	case map[string]any:
		for key, value := range item {
			if strings.EqualFold(key, "bot_token") {
				item[key] = maskToken(value)
				continue
			}
			maskSensitiveFields(value)
		}
	case []any:
		for _, value := range item {
			maskSensitiveFields(value)
		}
	}
}

func maskToken(v any) string {
	token, ok := v.(string)
	if !ok {
		return "***"
	}

	if len(token) <= 8 {
		return "***"
	}

	return token[:4] + "..." + token[len(token)-4:]
}

func truncateForLog(s string) string {
	if len(s) <= maxLoggedBodyBytes {
		return s
	}

	return s[:maxLoggedBodyBytes] + "...(truncated)"
}
