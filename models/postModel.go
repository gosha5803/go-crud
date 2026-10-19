package models

import (
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Title string `json:"title"`
	Body  string `json:"body"`
	AuthorID uint `gorm:"not null;index"`
	Views int `gorm:"not null;default:0" json:"views"`
}
