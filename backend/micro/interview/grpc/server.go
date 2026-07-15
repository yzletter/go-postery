package server

import (
	"context"

	interview_grpc "github.com/yzletter/go-postery/api/proto/interview/v1"
	"github.com/yzletter/go-postery/backend/micro/interview/service"
)

type InterviewServiceServer struct {
	svc service.InterviewService
	interview_grpc.UnimplementedInterviewServiceServer
}

func NewInterviewServiceServer(svc service.InterviewService) *InterviewServiceServer {
	return &InterviewServiceServer{
		svc: svc,
	}
}

func (server *InterviewServiceServer) Chat(ctx context.Context, req *interview_grpc.ChatRequest) (*interview_grpc.ChatResponse, error) {
	content, err := server.svc.Chat(ctx, req.UserID, req.Input)
	if err != nil {
		return &interview_grpc.ChatResponse{}, err
	}
	return &interview_grpc.ChatResponse{
		SessionID: req.SessionID,
		Content:   content,
	}, nil
}

func (server *InterviewServiceServer) StartInterview(ctx context.Context, req *interview_grpc.StartInterviewRequest) (*interview_grpc.StartInterviewResponse, error) {
	sessionID, err := server.svc.StartInterview(ctx, req.UserID, req.JD, req.Resume, req.CandidateName)
	if err != nil {
		return &interview_grpc.StartInterviewResponse{}, err
	}
	return &interview_grpc.StartInterviewResponse{
		SessionID: sessionID,
	}, nil
}

func (server *InterviewServiceServer) Answer(ctx context.Context, req *interview_grpc.AnswerRequest) (*interview_grpc.AnswerResponse, error) {
	if err := server.svc.Answer(ctx, req.UserID, req.SessionID, req.Answer); err != nil {
		return &interview_grpc.AnswerResponse{}, err
	}
	return &interview_grpc.AnswerResponse{}, nil
}

func (server *InterviewServiceServer) UploadQuestionsSign(ctx context.Context, req *interview_grpc.UploadQuestionsSignRequest) (*interview_grpc.UploadQuestionsSignResponse, error) {
	response, err := server.svc.UploadQuestionsSign(ctx, req.UserID)
	if err != nil {
		return &interview_grpc.UploadQuestionsSignResponse{}, err
	}
	return &interview_grpc.UploadQuestionsSignResponse{Response: response}, nil
}

func (server *InterviewServiceServer) UploadQuestionsCallback(ctx context.Context, req *interview_grpc.UploadQuestionsCallbackRequest) (*interview_grpc.UploadQuestionsCallbackResponse, error) {
	if err := server.svc.UploadQuestionsCallback(ctx, req.UserID, req.ObjectName); err != nil {
		return &interview_grpc.UploadQuestionsCallbackResponse{}, err
	}
	return &interview_grpc.UploadQuestionsCallbackResponse{}, nil
}

func (server *InterviewServiceServer) UploadQuestions(ctx context.Context, req *interview_grpc.UploadQuestionsRequest) (*interview_grpc.UploadQuestionsResponse, error) {
	if err := server.svc.UploadQuestions(ctx, req.UserID, req.SourceFile, req.Data); err != nil {
		return &interview_grpc.UploadQuestionsResponse{}, err
	}
	return &interview_grpc.UploadQuestionsResponse{}, nil
}

func (server *InterviewServiceServer) QuitInterview(ctx context.Context, req *interview_grpc.QuitInterviewRequest) (*interview_grpc.QuitInterviewResponse, error) {
	if err := server.svc.QuitInterview(ctx, req.UserID, req.SessionID); err != nil {
		return &interview_grpc.QuitInterviewResponse{}, err
	}
	return &interview_grpc.QuitInterviewResponse{}, nil
}

func (server *InterviewServiceServer) Evaluation(ctx context.Context, req *interview_grpc.EvaluationRequest) (*interview_grpc.EvaluationResponse, error) {
	if err := server.svc.Evaluation(ctx, req.UserID, req.SessionID); err != nil {
		return &interview_grpc.EvaluationResponse{}, err
	}
	return &interview_grpc.EvaluationResponse{}, nil
}

func (server *InterviewServiceServer) HealthCheck(ctx context.Context, req *interview_grpc.HealthCheckRequest) (*interview_grpc.HealthCheckResponse, error) {
	return &interview_grpc.HealthCheckResponse{}, nil
}
