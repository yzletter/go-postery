package manager

import (
	"context"
	"log/slog"
	"time"

	interview_grpc "github.com/yzletter/go-postery/api/proto/interview/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

const interviewRPCTimeout = 600000 * time.Second

type InterviewServiceManager struct {
	service string
	hub     ServiceHub
}

func NewInterviewManager(ctx context.Context, service string, hub ServiceHub) *InterviewServiceManager {
	hub.LoadEndpoints(ctx, service)
	hub.WatchEndpointsFromServiceHub(ctx, service)

	manager := &InterviewServiceManager{service: service, hub: hub}
	go manager.startHealthCheck(ctx) // 开启下游服务健康检查

	return manager
}

func (manager *InterviewServiceManager) Chat(ctx context.Context, req *interview_grpc.ChatRequest) (*interview_grpc.ChatResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interview_grpc.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc.ChatResponse
		resp, err = client.Chat(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *InterviewServiceManager) StartInterview(ctx context.Context, req *interview_grpc.StartInterviewRequest) (*interview_grpc.StartInterviewResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interview_grpc.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc.StartInterviewResponse
		resp, err = client.StartInterview(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *InterviewServiceManager) Answer(ctx context.Context, req *interview_grpc.AnswerRequest) (*interview_grpc.AnswerResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interview_grpc.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc.AnswerResponse
		resp, err = client.Answer(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *InterviewServiceManager) UploadQuestions(ctx context.Context, req *interview_grpc.UploadQuestionsRequest) (*interview_grpc.UploadQuestionsResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interview_grpc.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc.UploadQuestionsResponse
		resp, err = client.UploadQuestions(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *InterviewServiceManager) UploadQuestionsSign(ctx context.Context, req *interview_grpc.UploadQuestionsSignRequest) (*interview_grpc.UploadQuestionsSignResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interview_grpc.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interview_grpc.UploadQuestionsSignResponse
		resp, err = client.UploadQuestionsSign(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *InterviewServiceManager) UploadQuestionsCallback(ctx context.Context, req *interview_grpc.UploadQuestionsCallbackRequest) (*interview_grpc.UploadQuestionsCallbackResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interview_grpc.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc.UploadQuestionsCallbackResponse
		resp, err = client.UploadQuestionsCallback(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *InterviewServiceManager) QuitInterview(ctx context.Context, req *interview_grpc.QuitInterviewRequest) (*interview_grpc.QuitInterviewResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interview_grpc.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc.QuitInterviewResponse
		resp, err = client.QuitInterview(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *InterviewServiceManager) Evaluation(ctx context.Context, req *interview_grpc.EvaluationRequest) (*interview_grpc.EvaluationResponse, error) {
	var err = errs.ErrUnavailable
	var tryCnt = 1
	for try := 0; try < tryCnt; try++ {
		endpoint := manager.hub.Take(ctx, manager.service)
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}
		client := interview_grpc.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc.EvaluationResponse
		resp, err = client.Evaluation(ctx, req)
		cancel()

		if isEndpointFailure(err) {
			endpoint.MarkFailed()
			slog.Error("gRPC Error", "error", err, "service", manager.service, "endpoint", endpoint.Addr)
			continue
		}
		endpoint.MarkSuccess()
		return resp, err
	}

	return nil, err
}

func (manager *InterviewServiceManager) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			manager.checkOnce(ctx)
		}
	}
}

func (manager *InterviewServiceManager) checkOnce(ctx context.Context) {
	endpoints := manager.hub.GetEndpoints(ctx, manager.service)
	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}
		conn := endpoint.ClientConn()
		if conn == nil {
			continue
		}

		client := interview_grpc.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &interview_grpc.HealthCheckRequest{})
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
