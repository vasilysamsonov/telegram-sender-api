package request

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type SendMessage struct {
	BotToken string `json:"bot_token"`
	ChatID   ChatID `json:"chat_id"`
	Text     string `json:"text"`
}

type ChatID int64

func (c *ChatID) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		return fmt.Errorf("chat_id must not be empty")
	}

	var number int64
	if raw[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}

		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("chat_id must be a valid int64")
		}
		number = parsed
	} else {
		if err := json.Unmarshal(data, &number); err != nil {
			return err
		}
	}

	*c = ChatID(number)
	return nil
}
