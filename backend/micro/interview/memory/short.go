package memory

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/schema"
)

// ShortTermMemory 短期记忆：管理对话上下文的滑动窗口
type ShortTermMemory struct {
	mu     sync.RWMutex
	mp     map[int64][]*schema.Message // sessionID -> 消息列表
	maxLen int                         // 最大保留消息数
}

// NewShortTermMemory 创建短期记忆
func NewShortTermMemory(maxLen int) *ShortTermMemory {
	if maxLen <= 0 {
		maxLen = 20
	}
	return &ShortTermMemory{
		mp:     make(map[int64][]*schema.Message),
		maxLen: maxLen,
	}
}

// Add 添加消息到会话
func (m *ShortTermMemory) Add(_ context.Context, sessionID int64, msg *schema.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msgs := m.mp[sessionID]
	msgs = append(msgs, msg)

	// 滑动窗口：保留最近 maxLen 条消息
	if len(msgs) > m.maxLen {
		msgs = msgs[len(msgs)-m.maxLen:]
	}

	m.mp[sessionID] = msgs
}

// AddBatch 批量添加消息
func (m *ShortTermMemory) AddBatch(_ context.Context, sessionID int64, msgs []*schema.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing := m.mp[sessionID]
	existing = append(existing, msgs...)

	if len(existing) > m.maxLen {
		existing = existing[len(existing)-m.maxLen:]
	}

	m.mp[sessionID] = existing
}

// Get 获取会话的所有消息
func (m *ShortTermMemory) Get(_ context.Context, sessionID int64) []*schema.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	msgs := m.mp[sessionID]
	if len(msgs) == 0 {
		return nil
	}

	// 返回副本
	result := make([]*schema.Message, len(msgs))
	copy(result, msgs)
	return result
}

// GetRecent 获取最近 n 条消息
func (m *ShortTermMemory) GetRecent(_ context.Context, sessionID int64, n int) []*schema.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	msgs := m.mp[sessionID]
	if len(msgs) == 0 {
		return nil
	}

	start := 0
	if len(msgs) > n {
		start = len(msgs) - n
	}

	result := make([]*schema.Message, len(msgs)-start)
	copy(result, msgs[start:])
	return result
}

// Clear 清除会话记忆
func (m *ShortTermMemory) Clear(_ context.Context, sessionID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.mp, sessionID)
}
