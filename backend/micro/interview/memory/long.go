package memory

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/bytedance/sonic"
	"github.com/yzletter/go-postery/backend/micro/interview/model"
	"github.com/yzletter/go-postery/backend/micro/interview/repository"
)

const (
	maxAge = 30 * 24 * time.Hour // 薄弱点淘汰时间
	topN   = 10                  // 出题时只取最弱的 Top N
)

// LongTermMemory 长期记忆：管理用户画像和面试历史
type LongTermMemory struct {
	repo repository.InterviewRepository
}

// NewLongTermMemory 创建长期记忆
func NewLongTermMemory(repo repository.InterviewRepository) *LongTermMemory {
	return &LongTermMemory{
		repo: repo,
	}
}

func (memory *LongTermMemory) UpsertSession(ctx context.Context, userID int64, sessionID int64, state any) error {
	data, err := sonic.Marshal(state)
	if err != nil {
		return err
	}

	return memory.repo.UpsertSession(ctx, userID, sessionID, data)
}

// GetProfile 获取用户画像
func (memory *LongTermMemory) GetProfile(ctx context.Context, userID int64) (*model.InterviewProfile, error) {
	// 加载历史用户画像
	profile, err := memory.repo.LoadProfile(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
		return nil, err
	} else if profile != nil {
		return profile, nil
	}

	// 创建新用户画像
	profile = &model.InterviewProfile{
		UserID:     userID,
		SkillLevel: make(map[string]string),
	}

	return profile, nil
}

// UpdateWeakPoint 更新薄弱点
func (memory *LongTermMemory) UpdateWeakPoint(ctx context.Context, userID int64, topic string, score float64) error {
	profile, err := memory.GetProfile(ctx, userID)
	if err != nil {
		return err
	}

	// 查找已有薄弱点
	found := false
	for i := range profile.WeakPoints {
		if profile.WeakPoints[i].Topic == topic {
			profile.WeakPoints[i].Score = score
			profile.WeakPoints[i].HitCount++
			if score < 60 {
				profile.WeakPoints[i].WrongCount++
			}
			profile.WeakPoints[i].LastSeen = time.Now()
			found = true

			// 得分 >= 80 说明已掌握，移除薄弱点
			if score >= 80 {
				profile.WeakPoints = append(profile.WeakPoints[:i], profile.WeakPoints[i+1:]...)
			}
			break
		}
	}

	// 得分 < 60 才记录为薄弱点
	if !found && score < 60 {
		profile.WeakPoints = append(profile.WeakPoints, model.WeakPoint{
			Topic:      topic,
			Score:      score,
			HitCount:   1,
			WrongCount: 1,
			LastSeen:   time.Now(),
		})
	}

	// 落库
	return memory.repo.SaveProfile(ctx, profile)
}

// AddInterviewRecord 添加面试记录
func (memory *LongTermMemory) AddInterviewRecord(ctx context.Context, userID int64, record model.InterviewRecord) error {
	profile, err := memory.GetProfile(ctx, userID)
	if err != nil {
		return err
	}

	// 修改面试历史
	profile.InterviewHist = append(profile.InterviewHist, record)

	// 落库
	return memory.repo.SaveProfile(ctx, profile)
}

// GetWeakPoints 获取用户的薄弱点（淘汰过期 + 按分数排序 + Top N）
func (memory *LongTermMemory) GetWeakPoints(ctx context.Context, userID int64) []model.WeakPoint {
	profile, err := memory.GetProfile(ctx, userID)
	if err != nil || profile == nil {
		return nil
	}

	now := time.Now()

	// 过滤掉超过 30 天的过期薄弱点
	var active []model.WeakPoint
	for _, wp := range profile.WeakPoints {
		if now.Sub(wp.LastSeen) <= maxAge {
			active = append(active, wp)
		}
	}

	// 按分数升序排序（分数越低越弱，优先返回）
	sort.Slice(active, func(i, j int) bool {
		return active[i].Score < active[j].Score
	})

	// 只返回 Top N
	if len(active) > topN {
		active = active[:topN]
	}

	return active
}
