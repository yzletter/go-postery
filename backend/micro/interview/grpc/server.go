package server

import (
	"context"

	interview_grpc2 "github.com/yzletter/go-postery/backend/api/proto/interview/v1"
	"github.com/yzletter/go-postery/backend/micro/interview/service"
)

type InterviewServiceServer struct {
	svc service.InterviewService
	interview_grpc2.UnimplementedInterviewServiceServer
}

func NewInterviewServiceServer(svc service.InterviewService) *InterviewServiceServer {
	return &InterviewServiceServer{
		svc: svc,
	}
}

func (server *InterviewServiceServer) Chat(ctx context.Context, req *interview_grpc2.ChatRequest) (*interview_grpc2.ChatResponse, error) {
	content, err := server.svc.Chat(ctx, req.UserID, req.Input)
	if err != nil {
		return &interview_grpc2.ChatResponse{}, err
	}
	return &interview_grpc2.ChatResponse{
		SessionID: req.SessionID,
		Content:   content,
	}, nil
}

func (server *InterviewServiceServer) StartInterview(ctx context.Context, req *interview_grpc2.StartInterviewRequest) (*interview_grpc2.StartInterviewResponse, error) {
	sessionID, err := server.svc.StartInterview(ctx, req.UserID, req.JD, req.Resume, req.CandidateName)
	if err != nil {
		return &interview_grpc2.StartInterviewResponse{}, err
	}
	return &interview_grpc2.StartInterviewResponse{
		SessionID: sessionID,
	}, nil
}

func (server *InterviewServiceServer) Answer(ctx context.Context, req *interview_grpc2.AnswerRequest) (*interview_grpc2.AnswerResponse, error) {
	if err := server.svc.Answer(ctx, req.UserID, req.SessionID, req.Answer); err != nil {
		return &interview_grpc2.AnswerResponse{}, err
	}
	return &interview_grpc2.AnswerResponse{}, nil
}

func (server *InterviewServiceServer) UploadQuestionsSign(ctx context.Context, req *interview_grpc2.UploadQuestionsSignRequest) (*interview_grpc2.UploadQuestionsSignResponse, error) {
	response, err := server.svc.UploadQuestionsSign(ctx, req.UserID)
	if err != nil {
		return &interview_grpc2.UploadQuestionsSignResponse{}, err
	}
	return &interview_grpc2.UploadQuestionsSignResponse{Response: response}, nil
}

func (server *InterviewServiceServer) UploadQuestionsCallback(ctx context.Context, req *interview_grpc2.UploadQuestionsCallbackRequest) (*interview_grpc2.UploadQuestionsCallbackResponse, error) {
	if err := server.svc.UploadQuestionsCallback(ctx, req.UserID, req.ObjectName); err != nil {
		return &interview_grpc2.UploadQuestionsCallbackResponse{}, err
	}
	return &interview_grpc2.UploadQuestionsCallbackResponse{}, nil
}

func (server *InterviewServiceServer) UploadQuestions(ctx context.Context, req *interview_grpc2.UploadQuestionsRequest) (*interview_grpc2.UploadQuestionsResponse, error) {
	if err := server.svc.UploadQuestions(ctx, req.UserID, req.SourceFile, req.Data); err != nil {
		return &interview_grpc2.UploadQuestionsResponse{}, err
	}
	return &interview_grpc2.UploadQuestionsResponse{}, nil
}

func (server *InterviewServiceServer) QuitInterview(ctx context.Context, req *interview_grpc2.QuitInterviewRequest) (*interview_grpc2.QuitInterviewResponse, error) {
	if err := server.svc.QuitInterview(ctx, req.UserID, req.SessionID); err != nil {
		return &interview_grpc2.QuitInterviewResponse{}, err
	}
	return &interview_grpc2.QuitInterviewResponse{}, nil
}

func (server *InterviewServiceServer) Evaluation(ctx context.Context, req *interview_grpc2.EvaluationRequest) (*interview_grpc2.EvaluationResponse, error) {
	if err := server.svc.Evaluation(ctx, req.UserID, req.SessionID); err != nil {
		return &interview_grpc2.EvaluationResponse{}, err
	}
	return &interview_grpc2.EvaluationResponse{}, nil
}

func (server *InterviewServiceServer) HealthCheck(ctx context.Context, req *interview_grpc2.HealthCheckRequest) (*interview_grpc2.HealthCheckResponse, error) {
	return &interview_grpc2.HealthCheckResponse{}, nil
}
