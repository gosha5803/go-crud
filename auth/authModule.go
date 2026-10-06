package auth

import (
	"context"
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
	mailsQueue     *MailQueue
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
	appPort := os.Getenv("APP_PORT")
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

func getIntEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		log.Fatalf("env %s: ожидалось целое число, получено %q", key, raw)
	}
	return v
}

// TODO конфиг сервис
func loadMailQueueConfig() MailQueueConfig {
	return MailQueueConfig{
		Workers:    getIntEnv("MAIL_WORKERS", 3),
		QueueSize:  getIntEnv("MAIL_QUEUE_SIZE", 10),
		Retries:    getIntEnv("MAIL_ATTEMPTS", 3),
		RetryDelay: getDuration("MAIL_BASE_BACKOFF", 3*time.Second),
		MaxDelay:   getDuration("MAIL_MAX_RETRY_DELAY", 30*time.Second),
	}

}

func loadAuthConfig() AuthConfig {
	return AuthConfig{
		VerificationTokenTTL: getDuration("VERIFICATION_TOKEN_TTL", time.Hour*12),
		BcryptCost:           getIntEnv("BCRYPT_COST", 3),
	}

}

func NewAuthModule(g *gin.Engine, DB *gorm.DB) *AuthModule {
	mailServiceCfg := loadMailConfig()
	mailService := NewMailService(mailServiceCfg)

	// TODO userService пока не нужен контроллер и модуль свой и репы нет, пока его прям тут инстанциирую и внедрю
	userService := user.NewUserService(DB)

	mailQueueCfg := loadMailQueueConfig()
	mailQueue := NewMailQueue(mailQueueCfg, mailService)

	authConfig := loadAuthConfig()
	authService := NewAuthService(DB, mailQueue, authConfig, userService)
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
