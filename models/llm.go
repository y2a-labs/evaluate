package models

type LLM struct {
	Name string `db:"name" json:"name"`
	Url  string `db:"url" json:"url"`
}

type LLMModel struct {
	BaseModel
	LLM
}
