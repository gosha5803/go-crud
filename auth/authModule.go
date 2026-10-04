package auth

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosha5803/go-crud/user"
	"gorm.io/gorm"
)

type IAuthController interface {
	AssignRoutes()
}

type AuthModule struct {
	authController IAuthController
}

// TODO не уверен
// Типо можно в фонфиг сервисе передавать дженриком парсер функцию с фолбеком  и паникой!
func getDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Fatalf("env %s: ожидался duration (например, 24h), получено %q", key, raw)
	}
	return d
}

func loadMailConfig() MailConfig {
	// TODO нужен общий конфиг сервис с логикуой валидации .env
	appHost := os.Getenv("APP_HOST")
	appPort := os.Getenv("PORT")
	username := os.Getenv("SMTP_USERNAME")
	mailHost := os.Getenv("SMTP_HOST")
	password := os.Getenv("SMTP_PASSWORD")

	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))

	if err != nil {
		log.Fatal("NewAuthModule: loadMailConfig: cant convert password")
	}

	return MailConfig{
		appHost:      appHost,
		appPort:      appPort,
		mailUserName: username,
		mailHost:     mailHost,
		mailPassword: password,
		mailPort:     port,
	}

}

func NewAuthModule(g *gin.Engine, DB *gorm.DB) *AuthModule {
	mailServiceConfig := loadMailConfig()
	mailService := NewMailService(mailServiceConfig)

	verificationTokenTTL := getDuration("VERIFICATION_TOKEN_TTL", 24*time.Hour)
	bcryptCost, err := strconv.Atoi(os.Getenv("BCRYPT_COST"))

	if err != nil {
		bcryptCost = 12
		log.Fatal("NewAuthModule: failed to parse bcryptCost")
	}

	// TODO userService пока не нужен контроллер и модуль свой и репы нет, пока его прям тут инстанциирую и внедрю

	userService := user.NewUserService(DB)
	authService := NewAuthService(
		DB,
		mailService,
		AuthConfig{
			VerificationTokenTTL: verificationTokenTTL,
			BcryptCost:           bcryptCost,
		},
		userService,
	)
	authController := NewAuthController(g, "/auth", authService)

	return &AuthModule{
		authController: authController,
	}
}

func (m *AuthModule) Init() {
	m.authController.AssignRoutes()
}
