package posts

import (
	"errors"

	"gorm.io/gorm"
)

type CreatePostDto struct {
	Title string `json:"title" binding:"required,min=3,max=255"`
	Body  string `json:"body" binding:"required,min=1,max=10000"`
}

type UpdatePostDto struct {
	Title string `json:"title" binding:"min=3,max=255"`
	Body  string `json:"body" binding:"min=1,max=10000"`
}

type PostService struct {
	DB *gorm.DB
}

type IPostController interface {
	AssignRoutes()
}

var ErrPostNotFound = errors.New("post not found")
var ErrCouldNotDeletePost = errors.New("could not delete post")

// По NEST у нас должен быть сервис и контроллер
// В конструктор контроллера внедряется сервис
// Обозначаем интерфейс сервиса, потом структуру сервиса и инстанциируем его на уровне модуля.
//
