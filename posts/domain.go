package posts

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreatePostDto struct {
	Title string
	Body  string
}

type PostService struct {
	DB *gorm.DB
}

type IPostController interface {
	AssignRoutes()
	getPosts(c *gin.Context)
	getPostById(c *gin.Context)
	deletePost(c *gin.Context)
	createPost(c *gin.Context)
	updatePost(c *gin.Context)
}

var ErrPostNotFound = errors.New("post not found")
var ErrCouldNotDeletePost = errors.New("could not delete post")

// По NEST у нас должен быть сервис и контроллер
// В конструктор контроллера внедряется сервис
// Обозначаем интерфейс сервиса, потом структуру сервиса и инстанциируем его на уровне модуля.
//
