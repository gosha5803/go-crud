package auth

import (
	"net/http"
	"time"

	"github.com/gosha5803/go-crud/user"
	"gorm.io/gorm"
)

type VerificationToken struct {
	gorm.Model

	UserID uint `gorm:"not null;index"`
	// TODO auth сущность знает о USER сущности, возможно это дообогащение должно быть на другом уровен, в адаптерах
	User user.User `gorm:"constraint:OnDelete:CASCADE"`

	Hash string `gorm:"not null;size:128;index"`

	ExpiresAt time.Time `gorm:"not null;index"`
	UsedAt    *time.Time
}

type AuthReqDto struct {
	Password string `json:"password" binding:"required,min=3,max=25"`
	Email    string `json:"email" binding:"required,email"`
}

type AuthResDto struct {
	Message string `json:"message"`
}

type ActivateUserDto struct {
	Token string `form:"token" binding:"required,min=32,max=128"`
}

type UserActivatedDto struct {
	Message string `json:"message"`
}

type AuthConfig struct {
	VerificationTokenTTL time.Duration
	BcryptCost           int
}

type MailConfig struct {
	appHost      string
	appPort      string
	mailUserName string
	mailHost     string
	mailPassword string
	mailPort     int
}

type HTTPErr struct {
	Status  int
	Message string
	Err     error
}

func (err *HTTPErr) Error() string { return err.Message }
func (err *HTTPErr) Unwrap() error { return err.Err }

// TODO завязка на http
var (
	ErrEmailAlreadyUsed         = &HTTPErr{Message: "Пользователь с данным email уже существует", Status: http.StatusConflict}
	ErrVerificationTokenInvalid = &HTTPErr{Message: "Ссылка для подтверждения email не действительна", Status: http.StatusGone}
	ErrCouldNotSendEmail        = &HTTPErr{Message: "Ошибка отправки письма с подтверждением. Попробуйте снова", Status: http.StatusServiceUnavailable}
)
