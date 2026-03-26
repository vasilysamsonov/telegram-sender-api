package usecase

import (
	"context"

	"telegram-sender-api/internal/entity"
)

type Message interface {
	Send(context.Context, string, entity.Message) error
}
