package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string `gorm:"primary_key" json:"id"`
	CreatedAt time.Time	`json:"created_at,omitempty"`
	UpdatedAt time.Time	`json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"-"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	base.ID = uuid.NewString()
	return
}
