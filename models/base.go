package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string `gorm:"primary_key" example:"uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `json:"-"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	base.ID = uuid.NewString()
	return
}
