package models

import (
	"github.com/gosha5803/go-crud/user"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Title string `json:"title"`
	Body  string `json:"body"`
	AuthorID uint `gorm:"not null;index"`
	Views int `gorm:"not null;default:0" json:"views"`
	Author user.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}
