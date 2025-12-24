package message

type Request struct {
	Type        string `json:"type"`
	UserID      int64  `json:"user_id,string"`
	SessionID   int64  `json:"session_id,string"`
	SessionType int    `json:"session_type,omitempty"`
	MessageFrom int64  `json:"message_from,string,omitempty"`
	MessageTo   int64  `json:"message_to,string,omitempty"`
	Content     string `json:"content,omitempty"`
}
