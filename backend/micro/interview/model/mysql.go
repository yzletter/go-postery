package model

import (
	"time"
)

// InterviewProfile 用户画像（长期记忆）
type InterviewProfile struct {
	UserID        int64             `gorm:"column:id;primaryKey" json:"user_id"`                   // 用户 ID
	Name          string            `gorm:"column:name" json:"name"`                               // 用户名称
	SkillLevel    map[string]string `gorm:"column:skill_level;serializer:json" json:"skill_level"` // 技能 -> 水平（beginner/intermediate/advanced）
	WeakPoints    []WeakPoint       `gorm:"column:weak_points;serializer:json" json:"weak_points"` // 薄弱点列表
	InterviewHist []InterviewRecord `gorm:"-" json:"interview_hist"`                               // 面试历史，由 interview_records 单独存储
	CreatedAt     time.Time         `gorm:"column:created_at" json:"created_at"`                   // 创建时间
	UpdatedAt     time.Time         `gorm:"column:updated_at" json:"updated_at"`                   // 更新时间
}

func (u InterviewProfile) TableName() string {
	return "interview_profiles"
}

// WeakPoint 薄弱知识点
type WeakPoint struct {
	Topic      string    `json:"topic"`
	Score      float64   `json:"score"`       // 最近一次得分
	HitCount   int       `json:"hit_count"`   // 被考察次数
	WrongCount int       `json:"wrong_count"` // 答错次数
	LastSeen   time.Time `json:"last_seen"`
}

// InterviewRecord 面试记录摘要
type InterviewRecord struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id,omitempty"`    // 记录 ID
	UserID         int64     `gorm:"column:user_id" json:"user_id"`                             // 用户 ID
	SessionID      int64     `gorm:"column:session_id" json:"session_id"`                       // 面试会话 ID
	Position       string    `gorm:"column:position" json:"position"`                           // 面试岗位
	OverallScore   float64   `gorm:"column:overall_score" json:"overall_score"`                 // 综合得分
	ReportJSON     string    `gorm:"column:report_json" json:"report_json,omitempty"`           // 面试报告 JSON
	ReviewPlanJSON string    `gorm:"column:review_plan_json" json:"review_plan_json,omitempty"` // 复习计划 JSON
	Date           time.Time `gorm:"-" json:"date"`                                             // 画像缓存中的历史日期
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`                       // 创建时间
}

func (s InterviewRecord) TableName() string {
	return "interview_records"
}

// InterviewSession 面试会话
type InterviewSession struct {
	ID          int64     `gorm:"column:id;primaryKey" json:"id"`          // 面试会话 ID
	UserID      int64     `gorm:"column:user_id" json:"user_id"`           // 用户 ID
	SessionData []byte    `gorm:"column:session_data" json:"session_data"` // 会话快照
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`     // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`     // 更新时间
}

func (s InterviewSession) TableName() string {
	return "interview_sessions"
}
