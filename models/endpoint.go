package models

import (
	"fmt"
	"gorm.io/gorm"
)

// Endpoint представляет собой запись в таблице настроек для хранения URL стендов
type Endpoint struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// BeforeCreate хук для валидации перед созданием
func (e *Endpoint) BeforeCreate(tx *gorm.DB) error {
	if e.Name == "" || e.URL == "" {
		return fmt.Errorf("name and URL cannot be empty")
	}
	return nil
}
