package db

import (
	"chiyogami/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("pastes.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to SQLite database")
	}
	DB.AutoMigrate(&models.Paste{}, &models.User{})
	DB.Logger = logger.Default.LogMode(logger.Silent) //Kill overly-verbose gorm db errors, generic handlers are enough
}
