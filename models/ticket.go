package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Ticket struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `gorm:"index" json:"deleted_at,omitempty"` // Изменено на time.Time
	UserID      uint      `json:"user_id"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Status      string    `json:"status" gorm:"default:'open'"`
	ShortID     string    `json:"short_id" gorm:"default:gen_random_uuid()"`
	ClosedAt    time.Time `json:"closed_at,omitempty"`
	ClosedBy    string    `json:"closed_by,omitempty"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	if t.ShortID == "" {
		t.ShortID = uuid.New().String()
	}
	return nil
}
