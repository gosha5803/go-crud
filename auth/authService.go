package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/gosha5803/go-crud/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IMailService interface {
	AddVerificationJob(string, string) error
}

type IUserService interface {
	SetUserEmailVerified(*gorm.DB, uint, bool) error
	IsEmailExist(string) (bool, error)
	CreateUser(string, string) (user.User, error)
}

type AuthService struct {
	DB          *gorm.DB
	mailQueue   IMailService
	authConfig  AuthConfig
	userService IUserService
}

func NewAuthService(DB *gorm.DB, mailQueue IMailService, authCfg AuthConfig, userService IUserService) *AuthService {
	return &AuthService{
		DB:          DB,
		mailQueue:   mailQueue,
		authConfig:  authCfg,
		userService: userService,
	}
}

func (s *AuthService) ActivateUser(verificationToken string) error {
	// Изменить UsedAt у токена и EmailVerified
	// Тут конкурентность запросов с одинаковым токеном.
	// Оба могу пройти проверки валидности токен аи дважды его изменить?
	// Нет, строка конкретного токена блокируется на момент UPDATE и
	// второй запрос либо не найдёт такую строку либо потдвердит пользователя
	// где used_at IS NULL, если до этого отработает первый запрос
	tokenHash := s.hashVerificationToken(verificationToken)
	now := time.Now()

	tokenUpdateErr := s.DB.Transaction(func(tx *gorm.DB) error {
		var token VerificationToken

		res := tx.Model(&token).
			Clauses(clause.Returning{}).
			Where("hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, now).
			Update("used_at", now)
			// Убрали конкурентность апдейта и все проверки протухлости и использования и соответствия хеша в одном запросе к БД

		if res.Error != nil {
			return res.Error
		}

		// Если условия поиска токена для обновления не выполнятся,
		// ошибки не будет но в rows affected будет 0
		if res.RowsAffected == 0 {
			return ErrVerificationTokenInvalid
		}

		return s.userService.SetUserEmailVerified(tx, token.UserID, true)

	})
	// обновление токена и флага активации пользователя
	// в разных операциях, опять же пользователь может остаться не
	// активированным, а токен протухнет
	// Решается ли это ретраем активации почты?

	if tokenUpdateErr != nil {
		return fmt.Errorf("AuthService: ActivateUser: %w", tokenUpdateErr)
	}

	// сервис пользователя

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
		return user.User{}, fmt.Errorf("AuthService: Auth: %w", createUserErr)
	}

	// TODO gorutine, rabbit + monolith
	// Ретраи типо, не пришло письмо? попробовать ещё раз (и таймер)
	// TODO Если в итоге воркер упадёт с ошибкой,
	// На форме регистрации, как только введён валидный email улетает запрос,
	// подтверждён ли пользак
	// И если не подтверждён, тогда пишем пользаку, хотите подтвердить?
	// Или насильно заставляем. Кнопку даём, по которой отправляем письмо.
	// пользователь не сможет ни повторно зарегаться по новой email уже есть,
	// ни получить ссылку
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

	if err := s.mailQueue.AddVerificationJob(email, verificationToken); err != nil {
		if delErr := s.DB.Delete(&tokenModel).Error; delErr != nil {
			log.Printf("AuthService: sendVerificationMail: rollback failed: %v", delErr)
		}
		return fmt.Errorf("AuthService: sendVerificationMail: send mail: %w", err)
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
