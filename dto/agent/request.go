package agent

type ChatAgentRequest struct {
	SessionID string `json:"session_id"`
	Query     string `json:"query"`
}
