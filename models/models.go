package models

import (
	"time"

	"gorm.io/gorm"
)

type Paste struct {
	gorm.Model
	Title       string `gorm:"unique;size:4"`
	Content     string `gorm:"type:text"`
	Visibility  string
	Expiration  *time.Time `json:"expiration"`
	IsEncrypted bool
	UserID      uint
	IsUserPaste bool
}

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
}
