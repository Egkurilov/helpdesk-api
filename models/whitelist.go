package models

import (
	"time"
)

type Whitelist struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TelegramID   string    `gorm:"not null;index:idx_telegram_from,unique" json:"telegram_id"` // Уникальный индекс
	Text         string    `gorm:"not null;default:''" json:"text"`
	From         string    `gorm:"not null;index:idx_telegram_from,unique" json:"from" binding:"required,oneof=dev ift psi prom"` // Уникальный индекс
	ChatID       int64     `gorm:"not null;default:0" json:"chat_id"`
	FirstName    string    `gorm:"not null;default:''" json:"first_name"` // изменено
	LastName     string    `gorm:"not null;default:''" json:"last_name"`  // если необходимо
	Username     string    `gorm:"not null;default:''" json:"username"`   // если необходимо
	LanguageCode string    `json:"language_code"`
	Permission   string    `gorm:"default:'pending'" json:"permission"` // "pending", "approve", "deny"
	CreatedAt    time.Time `json:"create_date"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
