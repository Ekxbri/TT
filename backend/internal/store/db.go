package store

import (
    "log"
    "ai-hub/backend/internal/models"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func NewDB(dsn string) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    if err := db.AutoMigrate(&models.APIKey{}, &models.Project{}, &models.Task{}, &models.Prompt{}, &models.PromptHistory{}); err != nil {
        log.Fatal(err)
    }
    return db
}
