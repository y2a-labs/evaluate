package database

import (
	"script_validation/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(db_name string) *gorm.DB {
	var err error
	DB, err = gorm.Open(sqlite.Open(db_name), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB.AutoMigrate(
		&models.Conversation{},
		&models.Message{},
		&models.Prompt{},
		&models.MessageEvaluation{},
		&models.MessageEvaluationResult{},
		&models.LLM{},
		&models.Evaluation{},
	)
	return DB
}
