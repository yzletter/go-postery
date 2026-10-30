package domain

// ConnType 连接类型
type ConnType string

const (
	BizSession   ConnType = "session_connection"
	BizInterview ConnType = "interview_connection"
)

type MessageBiz string
