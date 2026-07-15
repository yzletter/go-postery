package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	oss_grpc "github.com/yzletter/go-postery/api/proto/oss/v1"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/interview/dag"
	"github.com/yzletter/go-postery/backend/micro/interview/dag/node"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/loader"
	_model "github.com/yzletter/go-postery/backend/micro/interview/model"
	"github.com/yzletter/go-postery/backend/micro/interview/repository"
	"github.com/yzletter/go-postery/backend/micro/interview/skill"
	"github.com/yzletter/go-postery/backend/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type interviewService struct {
	questionParser  *loader.QuestionParser
	model           model.ToolCallingChatModel
	orchestrator    *dag.Orchestrator
	skillRegistry   *skill.SkillRegistry
	repo            repository.InterviewRepository
	idGen           ports.IDGenerator
	ossClient       manager.OSSClient
	prepareGraph    compose.Runnable[*node.RunState, *node.RunState]
	interviewGraph  compose.Runnable[*node.RunState, *node.RunState]
	evaluationGraph compose.Runnable[*node.RunState, *node.RunState]
}

func NewInterviewService(wsGatewayClient manager.WSGatewayClient, orchestrator *dag.Orchestrator, skillRegistry *skill.SkillRegistry, repo repository.InterviewRepository,
	chatModel model.ToolCallingChatModel,
	questionParser *loader.QuestionParser,
	ossClient manager.OSSClient,
	idGen ports.IDGenerator,
	prepareGraph compose.Runnable[*node.RunState, *node.RunState],
	interviewGraph compose.Runnable[*node.RunState, *node.RunState],
	evaluationGraph compose.Runnable[*node.RunState, *node.RunState]) InterviewService {
	if orchestrator != nil {
		orchestrator.Callbacks = NewInterviewCallback(wsGatewayClient)
	}
	return &interviewService{
		questionParser:  questionParser,
		model:           chatModel,
		orchestrator:    orchestrator,
		prepareGraph:    prepareGraph,
		interviewGraph:  interviewGraph,
		evaluationGraph: evaluationGraph,
		skillRegistry:   skillRegistry,
		repo:            repo,
		idGen:           idGen,
		ossClient:       ossClient,
	}
}

func (svc *interviewService) Chat(ctx context.Context, userID int64, input string) (string, error) {
	if userID == 0 || input == "" {
		return "", errs.ErrInvalidArgument
	}
	if svc.skillRegistry == nil {
		return "", status.Error(codes.Unimplemented, "interview chat not implemented")
	}
	matchedSkill := svc.skillRegistry.Match(input)
	if matchedSkill == nil {
		return "", status.Error(codes.Unimplemented, "interview chat not implemented")
	}
	state := skill.NewSkillState(matchedSkill.Name())
	state.UserID = userID
	resp, err := matchedSkill.Handle(ctx, input, state)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (svc *interviewService) StartInterview(ctx context.Context, userID int64, jd string, resume string, candidateName string) (int64, error) {
	if userID == 0 || jd == "" || resume == "" {
		return 0, errs.ErrInvalidArgument
	}

	now := time.Now()
	input := &node.RunState{
		ID:         svc.idGen.NextID(),
		UserID:     userID,
		JDText:     jd,
		ResumeText: resume,
		Resume: &domain.Resume{
			RawText: resume,
			Name:    candidateName,
		},
		Status:    domain.StatusInit,
		CreatedAt: now,
		UpdatedAt: now,
	}

	output, err := svc.prepareGraph.Invoke(ctx, input)
	if err != nil {
		return 0, err
	}

	// 生成第一题
	output, err = svc.interviewGraph.Invoke(ctx, output)
	if err != nil {
		return 0, err
	}

	return output.ID, nil
}

func (svc *interviewService) Answer(ctx context.Context, userID int64, sessionID int64, answer string) error {
	if userID == 0 || sessionID == 0 || answer == "" {
		return errs.ErrInvalidArgument
	}

	state, err := svc.loadSession(ctx, userID, sessionID)
	if err != nil {
		return err
	} else if state.InterviewState == nil || state.JDAnalysis == nil || state.Resume == nil {
		return errs.ErrInvalidArgument
	}

	state.InterviewState.CurrentAnswer = answer
	if state.InterviewState.Phase == domain.PhaseWaitingAnswer {
		state.InterviewState.Phase = domain.PhaseAnswerComing
	} else if state.InterviewState.Phase == domain.PhaseWaitingFollowUp {
		state.InterviewState.Phase = domain.PhaseFollowUpComing
	} else {
		// 非法状态转移
		return errs.ErrInvalidArgument
	}

	_, err = svc.interviewGraph.Invoke(ctx, state)
	if err != nil {
		return err
	}

	return nil
}

func (svc *interviewService) QuitInterview(ctx context.Context, userID int64, sessionID int64) error {
	if userID == 0 || sessionID == 0 {
		return errs.ErrInvalidArgument
	}

	state, err := svc.loadSession(ctx, userID, sessionID)
	if err != nil {
		return err
	} else if state.InterviewState == nil || state.JDAnalysis == nil || state.Resume == nil {
		return errs.ErrInvalidArgument
	}

	state.InterviewState.Phase = domain.PhaseUserQuit
	_, err = svc.interviewGraph.Invoke(ctx, state)
	if err != nil {
		return err
	}

	_, err = svc.evaluationGraph.Invoke(ctx, state)
	if err != nil {
		return err
	}
	return nil
}

func (svc *interviewService) Evaluation(ctx context.Context, userID int64, sessionID int64) error {
	if userID == 0 || sessionID == 0 {
		return errs.ErrInvalidArgument
	}
	if svc.evaluationGraph == nil {
		return errs.ErrInternal
	}

	state, err := svc.loadSession(ctx, userID, sessionID)
	if err != nil {
		return err
	} else if state.InterviewState == nil || state.JDAnalysis == nil || state.Resume == nil {
		return errs.ErrInvalidArgument
	}

	_, err = svc.evaluationGraph.Invoke(ctx, state)
	if err != nil {
		return err
	}
	return nil
}

// UploadQuestionsSign 获取上传题库 OSS 签名
func (svc *interviewService) UploadQuestionsSign(ctx context.Context, id int64) (string, error) {
	if id <= 0 {
		slog.Debug("sign avatar upload rejected: invalid user id")
		return "", errs.ErrInvalidArgument
	}
	if svc.ossClient == nil {
		return "", status.Error(codes.Unimplemented, "oss client not configured")
	}

	resp, err := svc.ossClient.SignUpload(ctx, &oss_grpc.SignUploadRequest{
		Biz:      2,
		UserID:   id,
		FileName: "",
	})
	if err != nil {
		slog.Error("sign avatar upload failed", "error", err)
		return "", errs.ErrInternal
	}
	return resp.Response, err
}

// UploadQuestionsCallback 处理 OSS 上传题库回调
func (svc *interviewService) UploadQuestionsCallback(ctx context.Context, id int64, object string) error {
	if id <= 0 {
		slog.Debug("avatar callback rejected: invalid user id")
		return errs.ErrInvalidArgument
	}
	if svc.ossClient == nil {
		return status.Error(codes.Unimplemented, "oss client not configured")
	}

	// 校验前缀
	prefix := "interviews/questions/" + strconv.FormatInt(id, 10) + "/"
	if object == "" || !strings.HasPrefix(object, prefix) {
		slog.Debug("avatar callback rejected: invalid object")
		return errs.ErrInvalidArgument
	}

	// 获取题库 URL
	resp, err := svc.ossClient.GetObjectURL(ctx, &oss_grpc.GetObjectURLRequest{ObjectName: object})
	if err != nil {
		return errs.ErrInternal
	}

	// 获取内容
	data, err := downloadByURL(ctx, resp.URL)
	if err != nil {
		return errs.ErrInternal
	}

	return svc.UploadQuestions(ctx, id, path.Base(object), data)
}

// UploadQuestions 上传用户题库
func (svc *interviewService) UploadQuestions(ctx context.Context, userID int64, sourceFile string, data []byte) error {
	if userID <= 0 || sourceFile == "" || len(data) == 0 {
		return errs.ErrInvalidArgument
	}
	if svc.questionParser == nil || svc.model == nil {
		return status.Error(codes.Unimplemented, "question parser not configured")
	}
	result, err := svc.questionParser.ParseQuestionBank(ctx, svc.model, string(data))
	if err != nil {
		slog.Error("upload questions failed", "error", err)
		return errs.ErrInternal
	}
	if result.Success == 0 {
		return nil
	}

	if svc.orchestrator.MilvusStore != nil {
		questions := make([]_model.Question, 0)
		for _, q := range result.Questions {
			qq := _model.Question{
				ID:         q.ID,
				Content:    q.Content,
				Type:       q.Type,
				Difficulty: q.Difficulty,
				Skills:     q.Skills,
				Reference:  q.Reference,
			}
			questions = append(questions, qq)
		}
		if err := svc.orchestrator.MilvusStore.LoadParsedQuestions(ctx, userID, sourceFile, questions); err != nil {
			slog.Warn("")
		}
	}

	if svc.orchestrator.BM25Manager != nil {
		docs := make([]*schema.Document, 0, len(result.Questions))
		for _, q := range result.Questions {
			docs = append(docs, &schema.Document{
				ID:      strconv.FormatInt(q.ID, 10),
				Content: q.Content + "\n参考答案：" + q.Reference,
			})
		}
		svc.orchestrator.BM25Manager.ReplaceDocuments(userID, docs)
	}

	return nil
}

func (svc *interviewService) loadSession(ctx context.Context, userID int64, sessionID int64) (*node.RunState, error) {
	if userID == 0 || sessionID == 0 {
		return nil, errs.ErrInvalidArgument
	}
	if svc.repo == nil {
		return nil, errs.ErrInternal
	}

	data, err := svc.repo.LoadSession(ctx, sessionID)
	if err != nil {
		return nil, errs.ErrInternal
	}

	state := &node.RunState{}
	if err = sonic.Unmarshal(data, state); err != nil {
		return nil, errs.ErrInternal
	}
	if state.ID != sessionID || state.UserID != userID {
		return nil, errs.ErrInvalidArgument
	}
	return state, nil
}

func downloadByURL(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download question file failed: status=%d", resp.StatusCode)
	}

	const maxSize = 20 << 25
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxSize {
		return nil, fmt.Errorf("question file too large")
	}
	return data, nil
}
