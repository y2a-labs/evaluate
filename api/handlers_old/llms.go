package apihandlers

import (
	"script_validation/models"
)

func (app *App) FindOrCreateLLMs(model_names []string) ([]models.LLM, error) {
	llms := make([]models.LLM, len(model_names))
	for i, name := range model_names {
		llm := models.LLM{ID: name}
		r := app.Db.FirstOrCreate(&llm)
		if r.Error != nil {
			return nil, r.Error
		}
		llms[i] = llm
	}
	return llms, nil
}
