package models

import "time"

type Provider struct {
	ID           string     `gorm:"primary_key" json:"id"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"`
	DeletedAt    *time.Time `json:"-"`
	BaseUrl      string     `gorm:"uniqueIndex"`
	EnvKey       string
	Requests     int
	Interval     int
	Unit         string
}
