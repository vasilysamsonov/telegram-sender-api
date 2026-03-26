package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"telegram-sender-api/internal/entity"
	"telegram-sender-api/internal/usecase/message"
	messageusecase "telegram-sender-api/internal/usecase/message"
	"telegram-sender-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type messageWebAPIStub struct {
	err error
}

func (s *messageWebAPIStub) SendMessage(_ context.Context, botToken string, message entity.Message) error {
	return s.err
}

func newTestApp(webapi *messageWebAPIStub) *fiber.App {
	app := fiber.New()
	NewRouter(app, message.New(webapi), logger.New("debug"))
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("boom")
	})
	return app
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	defer resp.Body.Close()
	if _, err := io.ReadAll(resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != expected {
		t.Fatalf("expected status %d, got %d", expected, resp.StatusCode)
	}
}

func assertRequestIDHeader(t *testing.T, resp *http.Response) {
	t.Helper()

	if resp.Header.Get("X-Request-Id") == "" && resp.Header.Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header")
	}
}

func TestSendMessageRoute_ReturnsOK(t *testing.T) {
	t.Parallel()

	app := newTestApp(&messageWebAPIStub{})
	req, err := http.NewRequest(http.MethodPost, "/v1/messages/send", strings.NewReader(`{"bot_token":"token","chat_id":1,"text":"hello"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}

	assertStatus(t, resp, http.StatusOK)
	assertRequestIDHeader(t, resp)
}

func TestSendMessageRoute_ReturnsOKWithStringChatID(t *testing.T) {
	t.Parallel()

	app := newTestApp(&messageWebAPIStub{})
	req, err := http.NewRequest(http.MethodPost, "/v1/messages/send", strings.NewReader(`{"bot_token":"token","chat_id":"-1002852649500","text":"hello"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}

	assertStatus(t, resp, http.StatusOK)
	assertRequestIDHeader(t, resp)
}

func TestHealthzRoute_ReturnsOK(t *testing.T) {
	t.Parallel()

	app := newTestApp(&messageWebAPIStub{})
	req, err := http.NewRequest(http.MethodGet, "/healthz", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}

	assertStatus(t, resp, http.StatusOK)
	assertRequestIDHeader(t, resp)
}

func TestSendMessageRoute_ReturnsBadRequestOnInvalidJSON(t *testing.T) {
	t.Parallel()

	app := newTestApp(&messageWebAPIStub{})
	req, err := http.NewRequest(http.MethodPost, "/v1/messages/send", strings.NewReader(`{"chat_id":`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}

	assertStatus(t, resp, http.StatusBadRequest)
	assertRequestIDHeader(t, resp)
}

func TestSendMessageRoute_ReturnsInternalServerErrorOnUnknownError(t *testing.T) {
	t.Parallel()

	app := newTestApp(&messageWebAPIStub{err: errors.New("telegram unavailable")})
	req, err := http.NewRequest(http.MethodPost, "/v1/messages/send", strings.NewReader(`{"bot_token":"token","chat_id":1,"text":"hello"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}

	assertStatus(t, resp, http.StatusInternalServerError)
	assertRequestIDHeader(t, resp)
}

func TestSendMessageRoute_ReturnsBadGatewayOnExternalError(t *testing.T) {
	t.Parallel()

	app := newTestApp(&messageWebAPIStub{err: messageusecase.ErrExternal})
	req, err := http.NewRequest(http.MethodPost, "/v1/messages/send", strings.NewReader(`{"bot_token":"token","chat_id":1,"text":"hello"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}

	assertStatus(t, resp, http.StatusBadGateway)
	assertRequestIDHeader(t, resp)
}

func TestRecoveryMiddleware_ReturnsInternalServerErrorOnPanic(t *testing.T) {
	t.Parallel()

	app := newTestApp(&messageWebAPIStub{})
	req, err := http.NewRequest(http.MethodGet, "/panic", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}

	assertStatus(t, resp, http.StatusInternalServerError)
	assertRequestIDHeader(t, resp)
}
