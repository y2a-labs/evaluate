package database

import (
	"script_validation/models"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(db_name string) *gorm.DB {
	var err error
	DB, err = gorm.Open(sqlite.Open(db_name), &gorm.Config{
		NowFunc: func() time.Time {
			ti, _ := time.LoadLocation("UTC")
			return time.Now().In(ti)
		},
	})
	if err != nil {
		panic("failed to connect database")
	}
	DB.AutoMigrate(
		&models.Conversation{},
		&models.Message{},
		&models.MessageEvaluation{},
		&models.MessageEvaluationResult{},
		&models.LLM{},
		&models.Provider{},
	)
	DB.Set("gorm:time_zone", "UTC")
	return DB
}
