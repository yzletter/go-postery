package handler

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
	interview_grpc "github.com/yzletter/go-postery/api/proto/interview/v1"
	session_grpc "github.com/yzletter/go-postery/api/proto/session/v1"
	session_dto "github.com/yzletter/go-postery/backend/bff/dto/session"
	"github.com/yzletter/go-postery/backend/bff/ws_gateway"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
)

// Handler 统一处理 WebSocket 中的面试和聊天消息。
type Handler struct {
	interviewClient manager.InterviewClient
	sessionClient   manager.SessionClient
}

// NewHandler 创建通用 WebSocket Handler。
func NewHandler(interviewClient manager.InterviewClient, sessionClient manager.SessionClient) *Handler {
	return &Handler{
		interviewClient: interviewClient,
		sessionClient:   sessionClient,
	}
}

// NewSessionConnection  启动与 WebSocket 同生命周期的 Session 消息消费任务
// WebSocket 断开后 ctx 会被取消，Session 微服务中的 RabbitMQ consumer 随之退出。
func (hdl *Handler) NewSessionConnection(ctx context.Context, userID int64) error {
	if hdl.sessionClient == nil {
		return errs.ErrUnavailable
	}

	_, err := hdl.sessionClient.NewConnection(ctx, &session_grpc.UserID{UserID: userID})
	return err
}

// HandleWSMessage 根据 biz_type 将消息分发到对应模块。
func (hdl *Handler) HandleWSMessage(ctx context.Context, userID int64, msg ws_gateway.WSMessage) error {
	switch msg.BizType {
	case ws_gateway.WSBizTypeSession, "message", "read_ack":
		return hdl.handleSessionMessage(ctx, userID, msg)
	case ws_gateway.WSBizTypeInterview, "start_interview", "answer", "cancel_interview":
		return hdl.handleInterviewMessage(ctx, userID, msg)
	default:
		slog.Warn("unknown websocket biz type", "userID", userID, "bizType", msg.BizType)
		return nil
	}
}

// handleSessionMessage 处理聊天消息和已读回执。
func (hdl *Handler) handleSessionMessage(ctx context.Context, userID int64, msg ws_gateway.WSMessage) error {
	if hdl.sessionClient == nil {
		return errs.ErrUnavailable
	}

	data, err := sonic.Marshal(msg.BizData)
	if err != nil {
		return errs.ErrInvalidArgument
	}

	var req session_dto.Request
	if err = sonic.Unmarshal(data, &req); err != nil {
		return errs.ErrInvalidArgument
	}
	// 兼容 biz_type 直接使用 message/read_ack 的旧信封。
	if req.Type == "" {
		req.Type = msg.BizType
	}

	sessionID, err := strconv.ParseInt(req.SessionID, 10, 64)
	if err != nil || sessionID <= 0 {
		return errs.ErrInvalidArgument
	}

	switch req.Type {
	case "read_ack":
		_, err = hdl.sessionClient.ClearUnread(ctx, &session_grpc.ClearUnreadRequest{
			UserID:    userID,
			SessionID: sessionID,
		})
		return err
	case "message":
		messageTo, parseErr := strconv.ParseInt(req.MessageTo, 10, 64)
		if parseErr != nil || messageTo <= 0 || strings.TrimSpace(req.Content) == "" {
			return errs.ErrInvalidArgument
		}

		// 消息发送者始终使用鉴权得到的 userID，不能信任客户端传入值。
		_, err = hdl.sessionClient.Chat(ctx, &session_grpc.ChatRequest{
			UserID: userID,
			Message: &session_grpc.Message{
				SessionID:   sessionID,
				SessionType: int32(req.SessionType),
				MessageFrom: userID,
				MessageTo:   messageTo,
				Content:     req.Content,
			},
		})
		return err
	default:
		slog.Warn("unknown session websocket message", "userID", userID, "type", req.Type)
		return nil
	}
}

// interviewRequest 是面试模块的 WebSocket 请求参数。
type interviewRequest struct {
	Type          string `json:"type"`
	SessionID     string `json:"session_id"`
	JD            string `json:"jd"`
	Resume        string `json:"resume"`
	CandidateName string `json:"candidate_name"`
	Answer        string `json:"answer"`
}

// handleInterviewMessage 处理开始面试、回答和取消面试。
func (hdl *Handler) handleInterviewMessage(ctx context.Context, userID int64, msg ws_gateway.WSMessage) error {
	if hdl.interviewClient == nil {
		return errs.ErrUnavailable
	}

	data, err := sonic.Marshal(msg.BizData)
	if err != nil {
		return errs.ErrInvalidArgument
	}

	var req interviewRequest
	if err = sonic.Unmarshal(data, &req); err != nil {
		return errs.ErrInvalidArgument
	}
	// 兼容 biz_type 直接使用具体面试动作的旧信封。
	if req.Type == "" {
		req.Type = msg.BizType
	}

	switch req.Type {
	case "start_interview":
		if strings.TrimSpace(req.JD) == "" || strings.TrimSpace(req.Resume) == "" {
			return errs.ErrInvalidArgument
		}

		// 面试准备耗时较长，放到独立协程执行，避免阻塞 WebSocket Reader。
		go func() {
			_, callErr := hdl.interviewClient.StartInterview(ctx, &interview_grpc.StartInterviewRequest{
				UserID:        userID,
				JD:            req.JD,
				Resume:        req.Resume,
				CandidateName: req.CandidateName,
			})
			if callErr != nil && ctx.Err() == nil {
				slog.Error("start interview failed", "userID", userID, "error", callErr)
			}
		}()
		return nil
	case "answer":
		sessionID, parseErr := strconv.ParseInt(req.SessionID, 10, 64)
		if parseErr != nil || sessionID <= 0 || strings.TrimSpace(req.Answer) == "" {
			return errs.ErrInvalidArgument
		}

		// 回答处理可能触发模型调用，同样避免阻塞 WebSocket Reader。
		go func() {
			_, callErr := hdl.interviewClient.Answer(ctx, &interview_grpc.AnswerRequest{
				UserID:    userID,
				SessionID: sessionID,
				Answer:    req.Answer,
			})
			if callErr != nil && ctx.Err() == nil {
				slog.Error("answer interview failed", "userID", userID, "sessionID", sessionID, "error", callErr)
			}
		}()
		return nil
	case "cancel_interview":
		sessionID, parseErr := strconv.ParseInt(req.SessionID, 10, 64)
		if parseErr != nil || sessionID <= 0 {
			return errs.ErrInvalidArgument
		}

		_, err = hdl.interviewClient.QuitInterview(ctx, &interview_grpc.QuitInterviewRequest{
			UserID:    userID,
			SessionID: sessionID,
		})
		return err
	default:
		slog.Warn("unknown interview websocket message", "userID", userID, "type", req.Type)
		return nil
	}
}
