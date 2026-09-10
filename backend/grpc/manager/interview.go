package manager

import (
	"context"
	"log/slog"
	"time"

	interview_grpc2 "github.com/yzletter/go-postery/backend/api/proto/interview/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
)

const interviewRPCTimeout = 30 * time.Minute

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

func (manager *InterviewServiceManager) Chat(ctx context.Context, req *interview_grpc2.ChatRequest) (*interview_grpc2.ChatResponse, error) {
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
		client := interview_grpc2.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc2.ChatResponse
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

func (manager *InterviewServiceManager) StartInterview(ctx context.Context, req *interview_grpc2.StartInterviewRequest) (*interview_grpc2.StartInterviewResponse, error) {
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
		client := interview_grpc2.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc2.StartInterviewResponse
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

func (manager *InterviewServiceManager) Answer(ctx context.Context, req *interview_grpc2.AnswerRequest) (*interview_grpc2.AnswerResponse, error) {
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
		client := interview_grpc2.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc2.AnswerResponse
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

func (manager *InterviewServiceManager) UploadQuestions(ctx context.Context, req *interview_grpc2.UploadQuestionsRequest) (*interview_grpc2.UploadQuestionsResponse, error) {
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
		client := interview_grpc2.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc2.UploadQuestionsResponse
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

func (manager *InterviewServiceManager) UploadQuestionsSign(ctx context.Context, req *interview_grpc2.UploadQuestionsSignRequest) (*interview_grpc2.UploadQuestionsSignResponse, error) {
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
		client := interview_grpc2.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		var resp *interview_grpc2.UploadQuestionsSignResponse
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

func (manager *InterviewServiceManager) UploadQuestionsCallback(ctx context.Context, req *interview_grpc2.UploadQuestionsCallbackRequest) (*interview_grpc2.UploadQuestionsCallbackResponse, error) {
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
		client := interview_grpc2.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc2.UploadQuestionsCallbackResponse
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

func (manager *InterviewServiceManager) QuitInterview(ctx context.Context, req *interview_grpc2.QuitInterviewRequest) (*interview_grpc2.QuitInterviewResponse, error) {
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
		client := interview_grpc2.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc2.QuitInterviewResponse
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

func (manager *InterviewServiceManager) Evaluation(ctx context.Context, req *interview_grpc2.EvaluationRequest) (*interview_grpc2.EvaluationResponse, error) {
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
		client := interview_grpc2.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, interviewRPCTimeout)
		var resp *interview_grpc2.EvaluationResponse
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

		client := interview_grpc2.NewInterviewServiceClient(conn)

		ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)
		_, err := client.HealthCheck(ctx, &interview_grpc2.HealthCheckRequest{})
		cancel()

		if err != nil {
			endpoint.MarkFailed()
			continue
		}
		endpoint.MarkSuccess()
	}
}
