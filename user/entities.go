package user

import (
	"errors"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Password      string `gorm:"not null;size:255"`
	Email         string `gorm:"uniqueIndex;not null;size:255"`
	EmailVerified bool   `gorm:"not null;default:false"`
}

var ErrUserNotFound = errors.New("User not found")