package models

import (
	"time"
)

type LLM struct {
	BaseModel
	DeletedAt  *time.Time `json:"-"`
	ProviderID string
	Provider   Provider
}

type LLMCreate struct {
	// TODO add ressources
	ID         string
	ProviderID string
}

type LLMUpdate struct {
	// TODO add ressources
	ID string `json:"id"`
}
