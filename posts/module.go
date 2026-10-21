package posts

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/auth"
	"github.com/gosha5803/go-crud/initializers"
	"gorm.io/gorm"
)

type PostModule struct {
	postController IPostController
}

func NewPostModule(
	gin *gin.Engine, 
	DB *gorm.DB,
	log *slog.Logger,
	authCfg *initializers.AuthConfig,
	) *PostModule {
	// Ни сервис ни контроллер ничего не знают друг о друге кроме типов
	repo := NewPostRepository(DB)
	service := NewPostService(repo, log)
	authMiddleware := auth.AuthMiddleware([]byte(authCfg.JwtTokenSecret), log)
	controller := NewPostController(gin, "/post", service, authMiddleware)

	return &PostModule{
		postController: controller,
	}

}

func (module *PostModule) Init() {
	module.postController.AssignRoutes()
}
