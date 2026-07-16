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
var DBPath string

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
	// _txlock=immediate prevents SQLITE_BUSY on concurrent writes by
	// acquiring the write lock at BEGIN time where busy_timeout is honored.
	// _busy_timeout=30000 gives headroom for large payloads under burst.
	dsn := dbPath + "?_txlock=immediate&_busy_timeout=30000"
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to SQLite database: ", err)
	}

	// Auto-migrate schema
	err = DB.AutoMigrate(&models.Paste{}, &models.User{})
	if err != nil {
		log.Fatal("Failed to auto-migrate schema: ", err)
	}

	// Silence verbose GORM logs
	DB.Logger = logger.Default.LogMode(logger.Silent)

	// Check file for health checks
	DBPath = dbPath
}
