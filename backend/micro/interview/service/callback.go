package service

import (
	"context"
	"strconv"

	"github.com/bytedance/sonic"
	ws_gateway_grpc "github.com/yzletter/go-postery/api/proto/ws_gateway/v1"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/dag/node"
)

type Callback struct {
	ws manager.WSGatewayClient
}

func NewInterviewCallback(websocketClient manager.WSGatewayClient) *Callback {
	return &Callback{
		ws: websocketClient,
	}
}

// ServerMsg 服务端消息
type ServerMsg struct {
	Type            string   `json:"type"`
	SessionID       string   `json:"session_id"`
	Content         string   `json:"content,omitempty"`
	Stage           string   `json:"stage,omitempty"`
	Message         string   `json:"message,omitempty"`
	QuestionNum     int      `json:"question_num,omitempty"`
	Score           float64  `json:"score,omitempty"`
	Feedback        string   `json:"feedback,omitempty"`
	KeyPointsHit    []string `json:"key_points_hit,omitempty"`
	KeyPointsMissed []string `json:"key_points_missed,omitempty"`
}

func (c *Callback) OnStageChange(ctx context.Context, userID int64, sessionID int64, stage node.DAGStage, msg string) {
	c.push(ctx, userID, sessionID, "interview_stage_change", ServerMsg{Type: "stage_change", Stage: string(stage), Message: msg})
}

func (c *Callback) OnQuestion(ctx context.Context, userID int64, sessionID int64, questionNum int, content string) {
	c.push(ctx, userID, sessionID, "interview_question", ServerMsg{Type: "question", QuestionNum: questionNum, Content: content})
}

func (c *Callback) OnScore(ctx context.Context, userID int64, sessionID int64, score *agent.AnswerScore) {
	c.push(ctx, userID, sessionID, "interview_score", ServerMsg{
		Type: "score", Score: score.Score, Feedback: score.Feedback,
		KeyPointsHit: score.KeyPointsHit, KeyPointsMissed: score.KeyPointsMissed,
	})
}

func (c *Callback) OnReport(ctx context.Context, userID int64, sessionID int64, report string) {
	c.push(ctx, userID, sessionID, "interview_report", ServerMsg{Type: "report", Content: report})
}

func (c *Callback) OnReviewPlan(ctx context.Context, userID int64, sessionID int64, plan string) {
	c.push(ctx, userID, sessionID, "interview_review_plan", ServerMsg{Type: "review_plan", Content: plan})
}

func (c *Callback) GetUserAnswer(ctx context.Context, userID int64) (string, error) {
	return "", nil
}

func (c *Callback) push(ctx context.Context, userID int64, sessionID int64, bizType string, msg ServerMsg) {
	if c.ws == nil {
		return
	}
	msg.SessionID = strconv.FormatInt(sessionID, 10)
	bizData, err := sonic.Marshal(msg)
	if err != nil {
		return
	}
	_, _ = c.ws.Push(ctx, &ws_gateway_grpc.PushRequest{
		UserID:  userID,
		ConnBiz: ws_gateway_grpc.ConnBiz_CONNCECTION_Biz_Interview,
		BizType: bizType,
		BizData: bizData,
	})
}
