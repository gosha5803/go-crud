package posts

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PostModule struct {
	postController IPostController
}

func NewPostModule(gin *gin.Engine, DB *gorm.DB) *PostModule {
	// Ни сервис ни контроллер ничего не знают друг о друге кроме типов
	repo := NewPostRepository(DB)
	service := NewPostService(repo)
	controller := NewPostController(gin, "/post", service)

	return &PostModule{
		postController: controller,
	}

}

func (module *PostModule) Init() {
	module.postController.AssignRoutes()
}
