package service

import (
	"context"
	"encoding/json"
	"testing"

	ws_gateway_grpc "github.com/yzletter/go-postery/api/proto/ws_gateway/v1"
	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/dag/node"
)

type recordingWSGatewayClient struct {
	requests []*ws_gateway_grpc.PushRequest
}

func (c *recordingWSGatewayClient) Push(_ context.Context, req *ws_gateway_grpc.PushRequest) (*ws_gateway_grpc.PushResponse, error) {
	c.requests = append(c.requests, req)
	return &ws_gateway_grpc.PushResponse{}, nil
}

func TestInterviewCallbacksIncludeStringSessionID(t *testing.T) {
	const (
		userID    int64 = 42
		sessionID int64 = 9007199254740993
	)

	client := &recordingWSGatewayClient{}
	callback := NewInterviewCallback(client)
	ctx := context.Background()

	callback.OnStageChange(ctx, userID, sessionID, node.StageInterviewStart, "starting")
	callback.OnQuestion(ctx, userID, sessionID, 1, "question")
	callback.OnScore(ctx, userID, sessionID, &agent.AnswerScore{Score: 88})
	callback.OnReport(ctx, userID, sessionID, "report")
	callback.OnReviewPlan(ctx, userID, sessionID, "plan")

	wantBizTypes := []string{
		"interview_stage_change",
		"interview_question",
		"interview_score",
		"interview_report",
		"interview_review_plan",
	}
	if len(client.requests) != len(wantBizTypes) {
		t.Fatalf("Push() request count = %d, want %d", len(client.requests), len(wantBizTypes))
	}

	for i, req := range client.requests {
		if req.UserID != userID {
			t.Errorf("request %d user_id = %d, want %d", i, req.UserID, userID)
		}
		if req.BizType != wantBizTypes[i] {
			t.Errorf("request %d biz_type = %q, want %q", i, req.BizType, wantBizTypes[i])
		}

		var payload map[string]any
		if err := json.Unmarshal(req.BizData, &payload); err != nil {
			t.Fatalf("request %d biz_data is not valid JSON: %v", i, err)
		}
		got, ok := payload["session_id"].(string)
		if !ok {
			t.Fatalf("request %d session_id type = %T, want string; JSON: %s", i, payload["session_id"], req.BizData)
		}
		if got != "9007199254740993" {
			t.Errorf("request %d session_id = %q, want %q", i, got, "9007199254740993")
		}
	}
}
