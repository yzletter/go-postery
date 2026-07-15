package domain

import "time"

type UserInterviewProfile struct {
}

// JDAnalysis JD 分析结果
type JDAnalysis struct {
	RawJD            string            `json:"raw_jd"`           // 原始 JD 文本
	Position         string            `json:"position"`         // 岗位名称
	Company          string            `json:"company"`          // 公司名称
	RequiredSkills   []Skill           `json:"required_skills"`  // 必须技能
	PreferredSkills  []Skill           `json:"preferred_skills"` // 加分技能
	ExperienceLevel  string            `json:"experience_level"` // 经验要求（junior/mid/senior）
	Responsibilities []string          `json:"responsibilities"` // 岗位职责
	KeyTopics        []string          `json:"key_topics"`       // 面试重点方向
	Extra            map[string]string `json:"extra,omitempty"`  // 扩展字段
}

const (
	CategoryLanguage  = "language"
	CategoryFramework = "framework"
	CategoryDatabase  = "database"
	CategoryCloud     = "cloud"
	CategoryOther     = "other"
)

const (
	ImportanceMust      = "must"
	ImportancePreferred = "preferred"
)

type Skill struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Importance string `json:"importance"`
}

// Resume 简历结构化数据
type Resume struct {
	RawText    string      `json:"raw_text"`   // 原始简历文本
	Name       string      `json:"name"`       // 姓名
	Education  []Education `json:"education"`  // 教育经历
	Experience []WorkExp   `json:"experience"` // 工作经历
	Skills     []string    `json:"skills"`     // 技能列表
	Projects   []Project   `json:"projects"`   // 项目经历
}

const (
	DegreeBachelor = "bachelor"
	DegreeMaster   = "master"
	DegreePhd      = "phd"
)

// Education 教育经历
type Education struct {
	School string `json:"school"`
	Degree string `json:"degree"`
	Major  string `json:"major"`
	Year   string `json:"year"`
}

// WorkExp 工作经历
type WorkExp struct {
	Company     string   `json:"company"`
	Title       string   `json:"title"`
	Duration    string   `json:"duration"`
	Description string   `json:"description"`
	TechStack   []string `json:"tech_stack"`
}

// Project 项目经历
type Project struct {
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Description string   `json:"description"`
	TechStack   []string `json:"tech_stack"`
	Highlights  []string `json:"highlights"`
}

// ResumeMatchResult 简历与 JD 匹配结果
type ResumeMatchResult struct {
	OverallScore float64      `json:"overall_score"` // 综合匹配分（0-100）
	SkillMatch   []SkillMatch `json:"skill_match"`   // 技能匹配详情
	Strengths    []string     `json:"strengths"`     // 匹配优势
	Weaknesses   []string     `json:"weaknesses"`    // 薄弱环节
	FocusAreas   []string     `json:"focus_areas"`   // 面试重点考察方向
	ResumeGaps   []string     `json:"resume_gaps"`   // 简历空白点（可深挖）
}

// SkillMatch 单项技能匹配
type SkillMatch struct {
	SkillName  string  `json:"skill_name"`
	Required   bool    `json:"required"`    // 是否必须技能
	Matched    bool    `json:"matched"`     // 是否匹配
	MatchScore float64 `json:"match_score"` // 匹配度（0-100）
	Evidence   string  `json:"evidence"`    // 匹配证据（来自简历的哪部分）
}

// QuestionDirection 出题方向
type QuestionDirection struct {
	Topic       string          `json:"topic"`             // 出题方向/考点（如"sync.Map并发安全"）
	Type        QuestionType    `json:"type"`              // 类型（basic/experience/design）
	Difficulty  DifficultyLevel `json:"difficulty"`        // 难度（easy/medium/hard）
	SearchQuery string          `json:"search_query"`      // 用于题库检索的关键词
	Skills      []string        `json:"skills"`            // 考察技能点
	Context     string          `json:"context,omitempty"` // 简历中相关上下文（experience 类必填）
}

// QuestionDirectionPlan 出题方向规划, 由出题方向组成
type QuestionDirectionPlan struct {
	Directions []QuestionDirection `json:"directions"`
}

// QuestionPlan 出题计划
type QuestionPlan struct {
	TotalQuestions int                  `json:"total_questions"` // 计划出题总数
	Distribution   QuestionDistribution `json:"distribution"`    // 题目分布
	Questions      []PlannedQuestion    `json:"questions"`       // 规划的题目列表
}

// QuestionDistribution 题目类型分布（由 LLM 根据 JD 和简历动态决定）
type QuestionDistribution struct {
	Basic      int `json:"basic"`      // 基础知识题数
	Experience int `json:"experience"` // 工作/实习/项目经历题数
	Design     int `json:"design"`     // 系统设计题数
}

// PlannedQuestion 规划的面试题
type PlannedQuestion struct {
	ID         int64           `json:"id"`         // 面试流程内题目编号
	Content    string          `json:"content"`    // 题目内容
	Type       QuestionType    `json:"type"`       // 类型（basic/experience/design）
	Difficulty DifficultyLevel `json:"difficulty"` // 难度（easy/medium/hard）
	Skills     []string        `json:"skills"`     // 考察技能点
	FollowUps  []string        `json:"follow_ups"` // 预设追问
	Reference  string          `json:"reference"`  // 参考答案要点
	Source     string          `json:"source"`     // 来源：题库原题ID 或 "llm"
}

const (
	QuestionSourceLLM = "llm"
)

type QuestionType string

const (
	QuestionTypeBasic      QuestionType = "basic"
	QuestionTypeExperience QuestionType = "experience"
	QuestionTypeDesign     QuestionType = "design"
)

// DifficultyLevel 难度等级
type DifficultyLevel string

const (
	DifficultyEasy   DifficultyLevel = "easy"
	DifficultyMedium DifficultyLevel = "medium"
	DifficultyHard   DifficultyLevel = "hard"
)

// InterviewState 面试状态
type InterviewState struct {
	SessionID          int64              `json:"session_id"`
	AskedQuestionIDs   map[int64]struct{} `json:"asked_question_ids"`
	Phase              InterviewPhase     `json:"phase"`                // 当前面试阶段
	CurrentQuestionNum int                `json:"current_question_num"` // 当前第几题
	CurrentQuestion    PlannedQuestion    `json:"current_question"`     // 当前问题
	CurrentAnswer      string             `json:"current_answer"`       // 当前回答
	CurrentFollowup    string             `json:"current_followup"`     // 当前追问回答
	TotalQuestions     int                `json:"total_questions"`      // 总题数
	CurrentDifficulty  DifficultyLevel    `json:"current_difficulty"`   // 当前难度
	ConsecutiveRight   int                `json:"consecutive_right"`    // 连续答对
	ConsecutiveWrong   int                `json:"consecutive_wrong"`    // 连续答错
	QAHistory          []QAPair           `json:"qa_history"`           // 问答历史
	CandidateProfile   string             `json:"candidate_profile"`    // 候选人动态画像（面试过程中实时更新）
}

type InterviewPhase string

const (
	PhaseInitDone        InterviewPhase = "init_done"
	PhaseWaitingAnswer   InterviewPhase = "waiting_answer"
	PhaseAnswerDone      InterviewPhase = "answer_done"
	PhaseAnswerComing    InterviewPhase = "answer_coming"
	PhaseWaitingFollowUp InterviewPhase = "waiting_follow_up"
	PhaseFollowUpDone    InterviewPhase = "follow_up_done"
	PhaseFollowUpComing  InterviewPhase = "follow_up_coming"
	PhaseUpdateWPDone    InterviewPhase = "update_weakpoints_done"
	PhaseCompleted       InterviewPhase = "completed"
	PhaseUserQuit        InterviewPhase = "user_quit"
)

// QAPair 单次问答记录
type QAPair struct {
	Question     PlannedQuestion `json:"question"`
	UserAnswer   string          `json:"user_answer"`
	Score        float64         `json:"score"`          // 本题得分（0-100）
	Feedback     string          `json:"feedback"`       // 即时反馈
	FollowUpUsed bool            `json:"follow_up_used"` // 是否进行了追问
}

// EvaluationReport 面试评估报告
type EvaluationReport struct {
	SessionID      int64              `json:"session_id"`
	CandidateName  string             `json:"candidate_name"`
	Position       string             `json:"position"`
	OverallScore   float64            `json:"overall_score"`   // 综合得分
	OverallLevel   string             `json:"overall_level"`   // 综合评级（A/B/C/D）
	DimensionScore map[string]float64 `json:"dimension_score"` // 各维度得分
	Strengths      []string           `json:"strengths"`       // 表现优秀的方面
	Weaknesses     []string           `json:"weaknesses"`      // 需要提升的方面
	DetailedReview []QuestionReview   `json:"detailed_review"` // 逐题点评
	Summary        string             `json:"summary"`         // 综合评语
	CreatedAt      time.Time          `json:"created_at"`
}

// AnswerScore 回答评分结果
type AnswerScore struct {
	Score           float64  `json:"score"`
	Feedback        string   `json:"feedback"`
	KeyPointsHit    []string `json:"key_points_hit"`
	KeyPointsMissed []string `json:"key_points_missed"`
	ShouldFollowUp  bool     `json:"should_follow_up"`
}

// QuestionReview 单题点评
type QuestionReview struct {
	QuestionContent string   `json:"question_content"`
	UserAnswer      string   `json:"user_answer"`
	Score           float64  `json:"score"`
	Comment         string   `json:"comment"`
	KeyPointsHit    []string `json:"key_points_hit"`    // 命中的知识点
	KeyPointsMissed []string `json:"key_points_missed"` // 遗漏的知识点
}

// ReviewPlan 复习计划
type ReviewPlan struct {
	SessionID int64       `json:"session_id"`
	WeakAreas []WeakArea  `json:"weak_areas"` // 薄弱领域
	StudyPlan []StudyItem `json:"study_plan"` // 学习计划
	Resources []Resource  `json:"resources"`  // 推荐资源
	CreatedAt time.Time   `json:"created_at"`
}

// WeakArea 薄弱领域
type WeakArea struct {
	Topic    string  `json:"topic"`
	Score    float64 `json:"score"`    // 该领域得分
	Priority string  `json:"priority"` // 优先级（high/medium/low）
}

// StudyItem 学习项
type StudyItem struct {
	Topic        string   `json:"topic"`
	Objective    string   `json:"objective"`     // 学习目标
	Actions      []string `json:"actions"`       // 具体行动
	TimeEstimate string   `json:"time_estimate"` // 预估时间
}

// Resource 推荐资源
type Resource struct {
	Title string `json:"title"`
	Type  string `json:"type"` // article/video/repo/book
	URL   string `json:"url"`
	Desc  string `json:"desc"`
}

// SessionStatus 会话状态
type SessionStatus string

const (
	StatusInit          SessionStatus = "init"           // 初始化
	StatusJDAnalyzed    SessionStatus = "jd_analyzed"    // JD已分析
	StatusResumeMatched SessionStatus = "resume_matched" // 简历已匹配
	StatusPlanned       SessionStatus = "planned"        // 已出题
	StatusInterviewing  SessionStatus = "interviewing"   // 面试中
	StatusTerminated    SessionStatus = "terminated"     // 用户主动终止
	StatusEvaluated     SessionStatus = "evaluated"      // 已评估
	StatusCompleted     SessionStatus = "completed"      // 已完成
)
