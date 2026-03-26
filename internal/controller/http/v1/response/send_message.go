package response

type SendMessage struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
