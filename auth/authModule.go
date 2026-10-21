package auth

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/user"
	"gorm.io/gorm"
)

type IAuthController interface {
	AssignRoutes()
}

type AuthModule struct {
	authController IAuthController
	mailsQueue     *MailQueue
}

func NewAuthModule(
		g *gin.Engine, 
		DB *gorm.DB, 
		appConfig *initializers.AppConfig,
		log *slog.Logger,
	) *AuthModule {
	mailService := NewMailService(
		&appConfig.ApplicationConfig, 
		&appConfig.SMTPConfig,
		log,
	)

	// TODO userService пока не нужен контроллер и модуль свой и репы нет, пока его прям тут инстанциирую и внедрю
	userService := user.NewUserService(DB)
	jwtService := NewJwtSevice(&appConfig.AuthConfig)
	mailQueue := NewMailQueue(&appConfig.MailQueueConfig, mailService)

	authService := NewAuthService(
		DB, 
		mailQueue, 
		&appConfig.AuthConfig, 
		userService, 
		log,
		jwtService,
	)
	authController := NewAuthController(g, "/auth", authService)

	return &AuthModule{
		authController: authController,
		mailsQueue:     mailQueue,
	}
}

func (m *AuthModule) Init() {
	m.authController.AssignRoutes()
}

func (m *AuthModule) Close(ctx context.Context) error {
	return m.mailsQueue.Close(ctx)
}
