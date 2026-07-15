package question_scheduler

import "github.com/yzletter/go-postery/backend/micro/interview/domain"

// DifficultyAdjustFunc 难度调整函数
type DifficultyAdjustFunc func(cur domain.DifficultyLevel, consequentWrong int, consequentRight int) domain.DifficultyLevel

// 单个阶段的题目类型和对应题数
type stage struct {
	questionType domain.QuestionType
	questionNum  int
}

// SchedulerSnapshot 调度器快照
type SchedulerSnapshot struct {
	First            bool // 是否初始化
	StageNow         int  // 当前阶段
	StageHasAsked    int  // 当前阶段询问
	DifficultyNow    domain.DifficultyLevel
	ConsecutiveRight int
	ConsecutiveWrong int
}

type QuestionScheduler struct {
	first            bool                                             // 是否完成第一次
	ConsequentWrong  int                                              // 连续答错
	ConsequentRight  int                                              // 连续答对
	stageHasAsked    int                                              // 当前阶段已经问过的题目数
	stageNow         int                                              // 当前阶段
	difficultyNow    domain.DifficultyLevel                           // 当前难度
	candidates       map[domain.QuestionType][]domain.PlannedQuestion // 按照类型分类的候选题目
	stages           []stage                                          // 不同阶段
	difficultyAdjust DifficultyAdjustFunc                             // 难度调节函数
	questions        *questionPool                                    // 当前问题池
}

// 8 条基础题、5 条项目经验题、2 条实践设计题
var stages = []stage{
	{domain.QuestionTypeBasic, 8},
	{domain.QuestionTypeExperience, 5},
	{domain.QuestionTypeDesign, 2},
}

func RecoverQuestionScheduler(questions []domain.PlannedQuestion, snapshot *SchedulerSnapshot, AskedQuestionIDs map[int64]struct{}, difficultyAdjust DifficultyAdjustFunc) *QuestionScheduler {
	candidates := make(map[domain.QuestionType][]domain.PlannedQuestion)
	for _, question := range questions {
		if _, exists := AskedQuestionIDs[question.ID]; !exists {
			candidates[question.Type] = append(candidates[question.Type], question)
		}
	}

	return &QuestionScheduler{
		first:            snapshot.First,
		ConsequentWrong:  snapshot.ConsecutiveWrong,
		ConsequentRight:  snapshot.ConsecutiveRight,
		stageHasAsked:    snapshot.StageHasAsked,
		stageNow:         snapshot.StageNow,
		difficultyNow:    snapshot.DifficultyNow,
		candidates:       candidates,
		stages:           stages,
		difficultyAdjust: difficultyAdjust,
		questions:        newQuestionPool(candidates[stages[snapshot.StageNow].questionType]), // 恢复池子
	}
}

func NewQuestionScheduler(questions []domain.PlannedQuestion, difficultyAdjust DifficultyAdjustFunc) *QuestionScheduler {
	candidates := make(map[domain.QuestionType][]domain.PlannedQuestion)
	for _, question := range questions {
		candidates[question.Type] = append(candidates[question.Type], question)
	}

	return &QuestionScheduler{
		difficultyNow:    domain.DifficultyMedium,
		candidates:       candidates,
		stages:           stages,
		difficultyAdjust: difficultyAdjust,
		questions:        nil,
	}
}

// Save 保存快照
func (sched *QuestionScheduler) Save() *SchedulerSnapshot {
	return &SchedulerSnapshot{
		First:            sched.first,
		StageNow:         sched.stageNow,
		StageHasAsked:    sched.stageHasAsked,
		DifficultyNow:    sched.difficultyNow,
		ConsecutiveRight: sched.ConsequentRight,
		ConsecutiveWrong: sched.ConsequentWrong,
	}
}

// Next 下一题
func (sched *QuestionScheduler) Next() (domain.PlannedQuestion, domain.DifficultyLevel, bool) {
	for {
		// nil || 未加载过 || 池子空了 || 该阶段问题数已经达标
		if sched.questions == nil || sched.questions.empty() || sched.stageHasAsked >= sched.stages[sched.stageNow].questionNum {
			// 尝试进入下一阶段
			if ok := sched.nextStage(); !ok {
				return domain.PlannedQuestion{}, "", true
			}
		}

		// 取题
		picked, ok := sched.questions.next(sched.difficultyNow)
		if !ok {
			sched.questions = nil // 当前阶段候选已耗尽，强制推进到下一阶段
			continue
		}
		sched.stageHasAsked++
		return picked, sched.difficultyNow, false
	}
}

// Record 记录分数
func (sched *QuestionScheduler) Record(score float64) {
	if score > 70 {
		sched.ConsequentRight++
		sched.ConsequentWrong = 0
	} else {
		sched.ConsequentRight = 0
		sched.ConsequentWrong++
	}

	// 调整难度
	sched.difficultyNow = sched.difficultyAdjust(sched.difficultyNow, sched.ConsequentWrong, sched.ConsequentRight)
}

// Total 一共问多少个问题
func (sched *QuestionScheduler) Total() int {
	cnt := 0
	for _, stage := range sched.stages {
		cnt += min(stage.questionNum, len(sched.candidates[stage.questionType]))
	}
	return cnt
}

// nextStage 进入下一个阶段
func (sched *QuestionScheduler) nextStage() bool {
	startStage := 0          // 如果是第一次加载, 从第 0 阶段开始遍历
	if sched.first == true { // 如果加载过, 从当前阶段的后一阶段开始遍历
		startStage = sched.stageNow + 1
	}

	// 遍历阶段
	for nextStage := startStage; nextStage < len(sched.stages); nextStage++ {
		candidates := sched.candidates[sched.stages[nextStage].questionType]
		if candidates == nil || len(candidates) == 0 {
			// 该阶段没有候选
			continue
		}

		// 重置问题池
		sched.questions = newQuestionPool(sched.candidates[sched.stages[nextStage].questionType])
		// 更新阶段
		sched.stageNow = nextStage
		// 重置难度
		sched.difficultyNow = domain.DifficultyMedium
		// 重置连续答对、答错
		sched.ConsequentWrong = 0
		sched.ConsequentRight = 0
		// 重置已经问过的问题数
		sched.stageHasAsked = 0
		sched.first = true
		return true
	}

	return false
}
