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

type AuthReqDto struct {
	Password string `json:"password" binding:"required,min=3,max=25"`
	Email    string `json:"email" binding:"required,email"`
}

type AuthResDto struct {
	UserID  uint   `json:"userId"`
	Message string `json:"message"`
}

var ErrEmailAlreadyUsed = errors.New("email already exists")
