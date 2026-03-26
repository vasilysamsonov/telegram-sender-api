package telegram

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"telegram-sender-api/internal/entity"
	"telegram-sender-api/internal/usecase/message"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestMessageWebAPI_SendMessage_ReturnsExternalErrorOnNonOKStatus(t *testing.T) {
	t.Parallel()

	client := New(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"ok":false,"description":"bad request"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	})

	err := client.SendMessage(context.Background(), "token", entity.Message{
		ChatID: 1,
		Text:   "hello",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, message.ErrExternal) {
		t.Fatalf("expected external error, got %v", err)
	}
}

func TestMessageWebAPI_SendMessage_ReturnsExternalErrorOnTelegramAPIError(t *testing.T) {
	t.Parallel()

	client := New(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":false,"description":"denied"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	})

	err := client.SendMessage(context.Background(), "token", entity.Message{
		ChatID: 1,
		Text:   "hello",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, message.ErrExternal) {
		t.Fatalf("expected external error, got %v", err)
	}
}
