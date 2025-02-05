package db

import (
	"chiyogami/models"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	var err error
	var dbPath string

	// Set db path
	if _, ok := os.LookupEnv("DOCKER_ENV"); ok {
		dbPath = "/pastes/pastes.db"
	} else {
		// Default to ./pastes/pastes.db
		dbPath = os.Getenv("DATABASE_PATH")
		if dbPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			dbPath = filepath.Join(cwd, "pastes", "pastes.db")
		}
	}

	// Get absolute paths
	dbPath, err = filepath.Abs(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create the dir
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(dbPath), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Connect to SQLite
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to SQLite database")
	}

	// Auto-migrate schema
	DB.AutoMigrate(&models.Paste{}, &models.User{})

	// Silence verbose GORM logs
	DB.Logger = logger.Default.LogMode(logger.Silent)
}
