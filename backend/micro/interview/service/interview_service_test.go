package service

import (
	"context"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/compose"
	"github.com/yzletter/go-postery/backend/micro/interview/dag/node"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	interviewmodel "github.com/yzletter/go-postery/backend/micro/interview/model"
)

type staticInterviewRepository struct {
	session []byte
}

func (r *staticInterviewRepository) SaveProfile(context.Context, *interviewmodel.InterviewProfile) error {
	return nil
}

func (r *staticInterviewRepository) LoadProfile(context.Context, int64) (*interviewmodel.InterviewProfile, error) {
	return nil, nil
}

func (r *staticInterviewRepository) UpsertSession(_ context.Context, _ int64, _ int64, data []byte) error {
	r.session = append(r.session[:0], data...)
	return nil
}

func (r *staticInterviewRepository) LoadSession(context.Context, int64) ([]byte, error) {
	return append([]byte(nil), r.session...), nil
}

func TestAnswerRunsEvaluationWhenInterviewCompletes(t *testing.T) {
	const (
		userID    int64 = 7
		sessionID int64 = 11
	)

	repo := repositoryWithState(t, &node.RunState{
		ID:         sessionID,
		UserID:     userID,
		JDAnalysis: &domain.JDAnalysis{},
		Resume:     &domain.Resume{},
		InterviewState: &domain.InterviewState{
			Phase: domain.PhaseWaitingAnswer,
		},
	})

	interviewCalls := 0
	interviewGraph := compileRunStateGraph(t, func(_ context.Context, state *node.RunState) (*node.RunState, error) {
		interviewCalls++
		if state.InterviewState.Phase != domain.PhaseAnswerComing {
			t.Errorf("interview phase = %q, want %q", state.InterviewState.Phase, domain.PhaseAnswerComing)
		}
		state.InterviewState.Phase = domain.PhaseCompleted
		return state, nil
	})

	evaluationCalls := 0
	var evaluated *node.RunState
	evaluationGraph := compileRunStateGraph(t, func(_ context.Context, state *node.RunState) (*node.RunState, error) {
		evaluationCalls++
		evaluated = state
		return state, nil
	})

	svc := &interviewService{
		repo:            repo,
		interviewGraph:  interviewGraph,
		evaluationGraph: evaluationGraph,
	}
	if err := svc.Answer(context.Background(), userID, sessionID, "my answer"); err != nil {
		t.Fatalf("Answer() error = %v", err)
	}
	if interviewCalls != 1 {
		t.Errorf("interview graph calls = %d, want 1", interviewCalls)
	}
	if evaluationCalls != 1 {
		t.Fatalf("evaluation graph calls = %d, want 1", evaluationCalls)
	}
	if evaluated.InterviewState.Phase != domain.PhaseCompleted {
		t.Errorf("evaluated phase = %q, want %q", evaluated.InterviewState.Phase, domain.PhaseCompleted)
	}
	if evaluated.InterviewState.CurrentAnswer != "my answer" {
		t.Errorf("evaluated answer = %q, want %q", evaluated.InterviewState.CurrentAnswer, "my answer")
	}
}

func TestQuitInterviewSkipsInterviewGraphAndEvaluatesExistingAnswers(t *testing.T) {
	const (
		userID    int64 = 8
		sessionID int64 = 12
	)

	repo := repositoryWithState(t, &node.RunState{
		ID:         sessionID,
		UserID:     userID,
		JDAnalysis: &domain.JDAnalysis{},
		Resume:     &domain.Resume{},
		InterviewState: &domain.InterviewState{
			Phase: domain.PhaseWaitingAnswer,
			QAHistory: []domain.QAPair{
				{UserAnswer: "an existing answer", Score: 75},
			},
		},
	})

	interviewCalls := 0
	interviewGraph := compileRunStateGraph(t, func(_ context.Context, state *node.RunState) (*node.RunState, error) {
		interviewCalls++
		return state, nil
	})

	evaluationCalls := 0
	var evaluated *node.RunState
	evaluationGraph := compileRunStateGraph(t, func(_ context.Context, state *node.RunState) (*node.RunState, error) {
		evaluationCalls++
		evaluated = state
		return state, nil
	})

	svc := &interviewService{
		repo:            repo,
		interviewGraph:  interviewGraph,
		evaluationGraph: evaluationGraph,
	}
	if err := svc.QuitInterview(context.Background(), userID, sessionID); err != nil {
		t.Fatalf("QuitInterview() error = %v", err)
	}
	if interviewCalls != 0 {
		t.Errorf("interview graph calls = %d, want 0", interviewCalls)
	}
	if evaluationCalls != 1 {
		t.Fatalf("evaluation graph calls = %d, want 1", evaluationCalls)
	}
	if !evaluated.UserTerminated {
		t.Error("evaluated state UserTerminated = false, want true")
	}
	if evaluated.InterviewState.Phase != domain.PhaseUserQuit {
		t.Errorf("evaluated phase = %q, want %q", evaluated.InterviewState.Phase, domain.PhaseUserQuit)
	}
	if evaluated.Status != domain.StatusTerminated {
		t.Errorf("evaluated status = %q, want %q", evaluated.Status, domain.StatusTerminated)
	}
	if len(evaluated.InterviewState.QAHistory) != 1 {
		t.Errorf("evaluated QA history length = %d, want 1", len(evaluated.InterviewState.QAHistory))
	}
}

func repositoryWithState(t *testing.T, state *node.RunState) *staticInterviewRepository {
	t.Helper()
	data, err := sonic.Marshal(state)
	if err != nil {
		t.Fatalf("marshal session state: %v", err)
	}
	return &staticInterviewRepository{session: data}
}

func compileRunStateGraph(t *testing.T, invoke func(context.Context, *node.RunState) (*node.RunState, error)) compose.Runnable[*node.RunState, *node.RunState] {
	t.Helper()
	graph := compose.NewGraph[*node.RunState, *node.RunState]()
	const nodeName = "test_node"
	if err := graph.AddLambdaNode(nodeName, compose.InvokableLambda(invoke)); err != nil {
		t.Fatalf("add test graph node: %v", err)
	}
	if err := graph.AddEdge(compose.START, nodeName); err != nil {
		t.Fatalf("add test graph start edge: %v", err)
	}
	if err := graph.AddEdge(nodeName, compose.END); err != nil {
		t.Fatalf("add test graph end edge: %v", err)
	}
	runnable, err := graph.Compile(context.Background())
	if err != nil {
		t.Fatalf("compile test graph: %v", err)
	}
	return runnable
}
