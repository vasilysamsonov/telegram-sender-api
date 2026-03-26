package message

import (
	"context"
	"fmt"
	"strings"

	"telegram-sender-api/internal/entity"
	"telegram-sender-api/internal/repo"
)

type UseCase struct {
	webAPI repo.MessageWebAPI
}

func New(webAPI repo.MessageWebAPI) *UseCase {
	return &UseCase{
		webAPI: webAPI,
	}
}

func (uc *UseCase) Send(ctx context.Context, botToken string, message entity.Message) error {
	if strings.TrimSpace(botToken) == "" {
		return fmt.Errorf("bot token is empty: %w", ErrInvalidInput)
	}
	if message.ChatID == 0 {
		return fmt.Errorf("chat id is empty: %w", ErrInvalidInput)
	}
	if strings.TrimSpace(message.Text) == "" {
		return fmt.Errorf("text is empty: %w", ErrInvalidInput)
	}

	if err := uc.webAPI.SendMessage(ctx, botToken, message); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
