package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"telegram-sender-api/internal/entity"
	messageusecase "telegram-sender-api/internal/usecase/message"
)

type MessageWebAPI struct {
	client *http.Client
}

func New(client *http.Client) *MessageWebAPI {
	return &MessageWebAPI{
		client: client,
	}
}

type sendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func (w *MessageWebAPI) SendMessage(ctx context.Context, botToken string, msg entity.Message) error {
	payload := sendMessageRequest{
		ChatID: msg.ChatID,
		Text:   msg.Text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request: %w", messageusecase.ErrExternal)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", messageusecase.ErrExternal)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", messageusecase.ErrExternal)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", messageusecase.ErrExternal)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram returned status %d: %s: %w", resp.StatusCode, string(respBody), messageusecase.ErrExternal)
	}

	var telegramResp sendMessageResponse
	if err = json.Unmarshal(respBody, &telegramResp); err != nil {
		return fmt.Errorf("decode response: %w", messageusecase.ErrExternal)
	}

	if !telegramResp.OK {
		return fmt.Errorf("telegram error: %s: %w", telegramResp.Description, messageusecase.ErrExternal)
	}

	return nil
}
