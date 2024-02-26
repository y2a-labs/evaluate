package models

import (
	"time"
)

type LLM struct {
	ID         string     `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time  `json:"created_at,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at,omitempty"`
	DeletedAt  *time.Time `json:"-"`
	ProviderID string
	Provider   Provider
}
