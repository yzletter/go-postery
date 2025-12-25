package message

type Request struct {
	Type        string `json:"type"`
	UserID      string `json:"user_id,string"`
	SessionID   string `json:"session_id,string"`
	SessionType int    `json:"session_type,omitempty"`
	MessageFrom string `json:"message_from,string,omitempty"`
	MessageTo   string `json:"message_to,string,omitempty"`
	Content     string `json:"content,omitempty"`
}
