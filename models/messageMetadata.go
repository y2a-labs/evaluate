package models

import (
	"gorm.io/datatypes"
)

type MessageMetadata struct {
	BaseModel
	MessageID        string
	StartLatencyMs   int
	EndLatencyMs     int
	OutputTokenCount int
	InputTokenCount  int
	Embedding        datatypes.JSONSlice[float64]
}

type MessageMetadataCreate struct {
	// TODO add ressources
	ID string `json:"id"`
}

type MessageMetadataUpdate struct {
	// TODO add ressources
	ID string `json:"id"`
}
