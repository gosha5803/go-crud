package posts

import (
	"errors"

	"gorm.io/gorm"
)

type CreatePostDto struct {
	Title string `json:"title" binding:"required,min=3,max=255"`
	Body  string `json:"body" binding:"omitempty,max=10000"`
}

type UpdatePostDto struct {
	// если пользователь не редактирует title, он присылает пустой TITLE, и без флага omitEmpty валидация заголовка по длине не пройдёт
	// Валидация идёт по порядку, omitempty пропустит только те теги, что после него, если поле не пришло
	Title string `json:"title" binding:"omitempty,min=3,max=255"`
	Body  string `json:"body" binding:"max=10000"`
}

type PostIdPathParam struct {
	// Тег `uri:"id"` говорит Gin, какой параметр пути сюда биндить.
	// binding:"required" гарантирует, что параметр есть.
	ID uint `uri:"id" binding:"required"`
}

type PostService struct {
	DB *gorm.DB
}

type IPostController interface {
	AssignRoutes()
}

// Используется только для внутренней логики, не отдаётся пользаку
var ErrPostNotFound = errors.New("post not found")

// По NEST у нас должен быть сервис и контроллер
// В конструктор контроллера внедряется сервис
// Обозначаем интерфейс сервиса, потом структуру сервиса и инстанциируем его на уровне модуля.
//
