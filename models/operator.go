package models

import (
	"time"
)

type Operator struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`
	Username  string     `gorm:"unique;not null" json:"username"`
	Password  string     `gorm:"not null" json:"password"` // Хеш пароля
	Role      string     `gorm:"not null;default:'operator'" json:"role"`
}
