package models

import "gorm.io/datatypes"

type MessageEvaluationResult struct {
	BaseModel
	Content             string
	LatencyMs           int
	Similarity          float64
	MessageEvaluationID string
	Embedding           datatypes.JSONSlice[float64] `"json:"-"`
}
