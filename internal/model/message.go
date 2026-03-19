package model

import "time"

const (
	MessageTypePrivate   = 1
	MessageTypeGroup     = 2
	MessageTypeHeartbeat = 3
)

type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      int       `gorm:"not null;index" json:"type"`
	Target    string    `gorm:"size:64;index" json:"target,omitempty"`
	From      string    `gorm:"size:64;not null;index" json:"from"`
	Content   string    `gorm:"type:text;not null" json:"content,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
