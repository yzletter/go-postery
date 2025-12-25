package session

import "time"

type UpdateUnreadRequest struct {
	Updates Updates
	Delta   int
}

type Updates struct {
	LastMessageID   int64     `gorm:"last_message_id"`
	LastMessage     string    `gorm:"last_message"`
	LastMessageTime time.Time `gorm:"updated_at"`
}
