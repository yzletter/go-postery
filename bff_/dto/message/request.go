package message

type Request struct {
	Type        string `json:"type"`
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id,omitempty"`
	SessionType int    `json:"session_type,omitempty"`
	MessageFrom string `json:"message_from,omitempty"`
	MessageTo   string `json:"message_to,omitempty"`
	Content     string `json:"content,omitempty"`
}
