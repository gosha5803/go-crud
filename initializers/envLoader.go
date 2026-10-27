package initializers

import (
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type DBConfig struct {
	DBPort string `envconfig:"DB_PORT" required:"true"`
	DBUser string `envconfig:"DB_USER" required:"true"`
	DBPassword string `envconfig:"DB_PASSWORD" required:"true"`
	DBName string `envconfig:"DB_NAME" required:"true"`

	// Тут довольно безопасно использовать дефолты
	DBMaxOpenConns int `envconfig:"DB_MAX_OPEN_CONNS" default:"15"`
	DBMaxIdleConns int `envconfig:"DB_MAX_IDLE_CONNS" default:"8"`
	DBConnIdleTime time.Duration `envconfig:"DB_CONN_IDLE_TIME" default:"10m"`
	DBConnLifetime time.Duration `envconfig:"DB_CONN_LIFE_TIME" default:"20m"`
}

type SMTPConfig struct {
	SMTPPassword string `envconfig:"SMTP_PASSWORD" required:"true"`
	SMTPPort int `envconfig:"SMTP_PORT" required:"true"`
	SMTPHost string `envconfig:"SMTP_HOST" required:"true"`
	SMTPUsername string `envconfig:"SMTP_USERNAME" required:"true"`
}

type ApplicationConfig struct {
	AppPort int `envconfig:"APP_PORT" default:"8080"`
	AppHost string `envconfig:"APP_HOST" required:"true"`
}

type MailQueueConfig struct {
	MQWorkers int `envconfig:"MAIL_WORKERS" default:"3"`
	MQAttempts int `envconfig:"MAIL_ATTEMPTS" default:"3"`
	MQQueueSize int `envconfig:"MAIL_QUEUE_SIZE" default:"50"`
	MQBaseBackoff time.Duration `envconfig:"MAIL_BASE_BACKOFF" default:"3s"`
	MQMaxRetryDelay time.Duration `envconfig:"MAIL_MAX_RETRY_DELAY" default:"30s"`
}

type AuthConfig struct {
	VerificationTokenTTL time.Duration `envconfig:"VERIFICATION_TOKEN_TTL" default:"1h"`
	BcryptCost int `envconfig:"BCRYPT_COST" default:"10"`
	JwtTokenTTL time.Duration `envconfig:"JWT_TOKEN_TTL" default:"6h"`
	JwtTokenSecret string `envconfig:"JWT_TOKEN_SECRET" required:"true"`
}

type AiConfig struct {
	GeminiApiKey string `envconfig:"GEMINI_API_KEY" required:"true"`
}

type AppConfig struct {
	DBConfig DBConfig
	SMTPConfig SMTPConfig
	ApplicationConfig ApplicationConfig
	MailQueueConfig MailQueueConfig
	AuthConfig AuthConfig
	AiConfig AiConfig
	// TODO можно ли типизировать енамом
	Enviroment string `envconfig:"ENVIROMENT" required:"true"`
}
// TODO как прокидывать такие переменные через докер? 
func MustLoadConfig() *AppConfig {
	cfg, err := LoadEnv()

	if err != nil {
		log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
	}

	return cfg
}

func LoadEnv() (*AppConfig, error) {
	err := godotenv.Load()

	if err != nil {
		return &AppConfig{}, fmt.Errorf("error while loading env: %w", err)
	}

	var cfg AppConfig

	if err := envconfig.Process("", &cfg); err != nil {
        return nil, err // вернёт ошибку, если DB_PORT и т.д. не заданы
    }
	
    return &cfg, nil
}