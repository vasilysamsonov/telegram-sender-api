package repo

import (
	"context"

	"telegram-sender-api/internal/entity"
)

type MessageWebAPI interface {
	SendMessage(ctx context.Context, botToken string, message entity.Message) error
}
