package models

import (
	"time"
)

type Message struct {
	ID         uint `gorm:"primaryKey"`
	SenderID   uint
	ReceiverID uint
	Content    string
	Encrypted  bool
	CreatedAt  time.Time
}
