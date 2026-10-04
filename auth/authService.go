package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gosha5803/go-crud/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IMailService interface {
	SendVerificationMail(string, string) error
}

type IUserService interface {
	SetUserEmailVerified(uint, bool) error
	IsEmailExist(string) (bool, error)
	CreateUser(string, string) (user.User, error)
}

type AuthService struct {
	DB          *gorm.DB
	mailService IMailService
	authConfig  AuthConfig
	userService IUserService
}

func NewAuthService(DB *gorm.DB, mailService IMailService, authCfg AuthConfig, userService IUserService) *AuthService {
	return &AuthService{
		DB:          DB,
		mailService: mailService,
		authConfig:  authCfg,
		userService: userService,
	}
}

func (s *AuthService) ActivateUser(verificationToken string) error {
	tokenHash := s.hashVerificationToken(verificationToken)

	var token VerificationToken

	// Что вот сейчас происходит, я передаю указатель на безымянную структуру?
	err := s.DB.
		Where("hash = ?", tokenHash).
		First(&token).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("AuthService: ActivateUser: token not found: , %w", ErrVerificationTokenInvalid)
		}

		return fmt.Errorf("AuthService: ActivateUser: %w", err)
	}

	if token.UsedAt != nil {
		return fmt.Errorf("AuthService: ActivateUser: token already used: %w", ErrVerificationTokenInvalid)
	}

	// TODO как пользаку подтвердить почту, если токен протух? Удалять пользователя?
	if token.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("AuthService: ActivateUser: token expired: %w", ErrVerificationTokenInvalid)
	}

	// Изменить UsedAt у токена и EmailVerified
	tokenUpdateErr := s.DB.Exec(`
			UPDATE verification_tokens
			SET used_at = ?, updated_at = ?
			WHERE id = ? AND used_at IS NULL
		`, time.Now(), time.Now(), token.ID).Error

	if tokenUpdateErr != nil {
		return fmt.Errorf("AuthService: ActivateUser: %w", tokenUpdateErr)
	}

	// сервис пользователя
	userUpdateErr := s.userService.SetUserEmailVerified(token.UserID, true)

	if userUpdateErr != nil {
		return fmt.Errorf("AuthService: ActivateUser: %w", userUpdateErr)
	}

	return nil
}

// TODO мб покрыть ActivateUser тестами.
// TODO далее внедрять JWT
func (s *AuthService) Auth(dto AuthReqDto) (user.User, error) {
	// TODO где-то я подменяю обраение к модели напрямую через обращение к сервису, но не вот тут, например,
	// где мне надо вернуть пустого пользака
	emailExist, err := s.userService.IsEmailExist(dto.Email)

	if err != nil {
		return user.User{}, fmt.Errorf("AuthService: Auth: %w", err)
	}

	if emailExist {
		return user.User{}, fmt.Errorf("AuthService: Auth: %w", ErrEmailAlreadyUsed)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(dto.Password),
		s.authConfig.BcryptCost,
	)

	if err != nil {
		// TODO можно сделать метод, getEmptyUser, чтобы вообще не зтянуть модель, попробую
		return user.User{}, fmt.Errorf("AuthService: Auth: %w", err)
	}

	createdUser, createUserErr := s.userService.CreateUser(dto.Email, string(hashedPassword))

	if createUserErr != nil {
		return user.User{}, fmt.Errorf("AuthService: Auth: %w", err)
	}

	// TODO gorutine, rabbit + monolith
	// Ретраи типо, не пришло письмо? попробовать ещё раз (и таймер)
	if err := s.sendVerificationMail(createdUser.Email, createdUser.ID); err != nil {
		return user.User{}, fmt.Errorf("AuthService: Auth: %w", err)
	}

	return createdUser, nil
}

func (s *AuthService) sendVerificationMail(email string, userID uint) error {
	// TODO тут и логика сохранения токена и отправки email
	// Возможно надо раскидать по методам, а также обеспечить консистентность. Сейчас я просто удалаю токен из БД при ошибке
	// отправки письма. Но мб тут надо внедрить outbox паттерн
	verificationToken, err := s.createVerificationToken()

	if err != nil {
		return fmt.Errorf("AuthService: sendVerificationMail: %w", err)
	}

	tokenHash := s.hashVerificationToken(verificationToken)

	tokenModel := VerificationToken{
		Hash:      tokenHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.authConfig.VerificationTokenTTL),
	}

	if err := s.DB.Create(&tokenModel).Error; err != nil {
		return fmt.Errorf("AuthService: sendVerificationMail: createToken: %w", err)
	}

	if err := s.mailService.SendVerificationMail(email, verificationToken); err != nil {
		if delErr := s.DB.Delete(&tokenModel).Error; delErr != nil {
			log.Printf("AuthService: sendVerificationMail: rollback failed: %v", delErr)
		}
		return fmt.Errorf("AuthService: sendVerificationMail: send mail: %w", ErrCouldNotSendEmail)
	}

	return nil
}

func (s *AuthService) hashVerificationToken(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}

func (s *AuthService) createVerificationToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("AuthService: createVerificationToken: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
