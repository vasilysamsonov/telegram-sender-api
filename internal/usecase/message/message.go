package message

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"telegram-sender-api/internal/entity"
	"telegram-sender-api/internal/repo"
)

const MaxTextLength = 4096

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
	if utf8.RuneCountInString(message.Text) > MaxTextLength {
		return fmt.Errorf("text exceeds %d characters: %w", MaxTextLength, ErrInvalidInput)
	}

	if err := uc.webAPI.SendMessage(ctx, botToken, message); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}
