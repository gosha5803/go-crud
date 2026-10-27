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

type EmailDto struct {
	Email    string `json:"email" binding:"required,email"`
}

type LoginReqDto struct {
	EmailDto
	Password string `json:"password" binding:"required"`
}

type LoginResDto struct {
	AccessToken string `json:"accessToken"`
}

type AuthReqDto struct {
	Password string `json:"password" binding:"required,min=3,max=25"`
	EmailDto
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

type UserIdCtxKey struct {}

var UserIdKey = UserIdCtxKey{}

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
	ErrMailQueueFull            = &HTTPErr{Message: "Сервис отправки писем перегружен. Попробуйте позже", Status: http.StatusServiceUnavailable}
	ErrMailQueueClosed          = &HTTPErr{Message: "Сервис временно не доступен. Попробуйте позже", Status: http.StatusServiceUnavailable}
	ErrEmailVerified = &HTTPErr{Message: "Email уже подтверждён", Status: http.StatusConflict}
	ErrInvalidCredentials = &HTTPErr{Message: "Неверный логин или пароль", Status: http.StatusBadRequest}
	ErrUserIsNotActivated = &HTTPErr{Message: "Необходимо подтвердить email", Status: http.StatusForbidden}
	ErrUnauthorized = &HTTPErr{Message: "Пользователь неавторизован", Status: http.StatusUnauthorized}
)

var verificationEmailSendDto = AuthResDto{Message: "На вашу почту было отправена ссылка для подтверждения. Подтвердите вашу почту в течение 24 часов."}