package models

import (
	"time"
)

type Message struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"` // Заменяем gorm.DeletedAt на *time.Time
	TicketID  uint       `gorm:"not null" json:"ticket_id"`
	Sender    string     `gorm:"not null" json:"sender"`
	Recipient string     `gorm:"not null" json:"recipient"`
	Content   string     `gorm:"not null" json:"content"`
	Timestamp time.Time  `gorm:"autoCreateTime" json:"timestamp"`
}
