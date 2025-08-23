package models

import (
	"time"

	"gorm.io/gorm"
)

type Paste struct {
	gorm.Model
	Title       string     `gorm:"uniqueIndex;size:4"`
	Content     string     `gorm:"type:text"`
	Visibility  string     `gorm:"index:vis_enc"`
	Expiration  *time.Time `gorm:"index" json:"Expiration"`
	IsEncrypted bool       `gorm:"index:vis_enc"`
	UserID      uint       `gorm:"index:user_paste"`
	IsUserPaste bool       `gorm:"index:user_paste"`
}

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex"`
	Password string
}
