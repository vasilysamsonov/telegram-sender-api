package message

import (
	"context"
	"errors"
	"strings"
	"testing"

	"telegram-sender-api/internal/entity"
)

type messageWebAPIStub struct {
	called   bool
	botToken string
	message  entity.Message
	err      error
}

func (s *messageWebAPIStub) SendMessage(_ context.Context, botToken string, message entity.Message) error {
	s.called = true
	s.botToken = botToken
	s.message = message
	return s.err
}

func TestUseCaseSend_ValidatesInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		botToken string
		message  entity.Message
	}{
		{
			name:    "missing bot token",
			message: entity.Message{ChatID: 1, Text: "hello"},
		},
		{
			name:     "missing chat id",
			botToken: "token",
			message:  entity.Message{Text: "hello"},
		},
		{
			name:     "missing text",
			botToken: "token",
			message:  entity.Message{ChatID: 1},
		},
		{
			name:     "text too long",
			botToken: "token",
			message: entity.Message{
				ChatID: 1,
				Text:   strings.Repeat("a", MaxTextLength+1),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			webapi := &messageWebAPIStub{}
			uc := New(webapi)

			if err := uc.Send(context.Background(), tt.botToken, tt.message); err == nil {
				t.Fatal("expected validation error")
			}
			if webapi.called {
				t.Fatal("webapi must not be called on invalid input")
			}
		})
	}
}

func TestUseCaseSend_AllowsTextAtTelegramLimit(t *testing.T) {
	t.Parallel()

	webapi := &messageWebAPIStub{}
	uc := New(webapi)
	message := entity.Message{
		ChatID: 1,
		Text:   strings.Repeat("a", MaxTextLength),
	}

	if err := uc.Send(context.Background(), "token", message); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !webapi.called {
		t.Fatal("expected webapi to be called")
	}
}

func TestUseCaseSend_CallsWebAPI(t *testing.T) {
	t.Parallel()

	webapi := &messageWebAPIStub{}
	uc := New(webapi)
	message := entity.Message{
		ChatID: 1,
		Text:   "hello",
	}

	if err := uc.Send(context.Background(), "token", message); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !webapi.called {
		t.Fatal("expected webapi to be called")
	}
	if webapi.botToken != "token" {
		t.Fatalf("unexpected token: %s", webapi.botToken)
	}
	if webapi.message != message {
		t.Fatalf("unexpected message: %+v", webapi.message)
	}
}

func TestUseCaseSend_PropagatesExternalError(t *testing.T) {
	t.Parallel()

	webapi := &messageWebAPIStub{
		err: ErrExternal,
	}
	uc := New(webapi)

	err := uc.Send(context.Background(), "token", entity.Message{
		ChatID: 1,
		Text:   "hello",
	})
	if !errors.Is(err, ErrExternal) {
		t.Fatalf("expected external error, got %v", err)
	}
}
