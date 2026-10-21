package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/lib/logger/sl"
	"github.com/gosha5803/go-crud/user"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IMailService interface {
	AddVerificationJob(string, string) error
}

type IJwtService interface {
	NewToken(user.User) (string, error)
}

type IUserService interface {
	SetUserEmailVerified(*gorm.DB, uint, bool) error
	FindUserByEmail(context.Context, string) (user.User, error)
	CreateUser(string, string) (user.User, error)
}

type AuthService struct {
	log *slog.Logger
	DB          *gorm.DB
	mailQueue   IMailService
	authConfig  *initializers.AuthConfig
	userService IUserService
	jwtService IJwtService
}

func NewAuthService(
		DB *gorm.DB, 
		mailQueue IMailService, 
		authCfg *initializers.AuthConfig, 
		userService IUserService,
		log *slog.Logger,
		jwtService IJwtService,
	) *AuthService {
		
	return &AuthService{
		DB:          DB,
		mailQueue:   mailQueue,
		authConfig:  authCfg,
		userService: userService,
		log: log,
		jwtService: jwtService,
	}
}

func (s *AuthService) Login(
	ctx context.Context, 
	email string, 
	password string,
	) (string, error) {
	op := "AuthService.Login"
	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("Trying ro login user")
	
	existingUser, err := s.userService.FindUserByEmail(ctx, email)

	if err != nil {
		log.Error("Failed to find user by email", sl.Err(err))
		
		if errors.Is(err, user.ErrUserNotFound) {
			return "", fmt.Errorf("%s, %w", op, ErrInvalidCredentials)
		}
		
		return "", fmt.Errorf("%s, %w", op, err)
	}
	
	if !existingUser.EmailVerified {
		log.Error("Failed to login user who is not verified", sl.Err(ErrUserIsNotActivated))
		return "", fmt.Errorf("%s, %w", op, ErrUserIsNotActivated)
	}

	// TODO брутфорс
	if err := s.comparePassword([]byte(existingUser.Password), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Error("Wrong password", sl.Err(err))
			return "", fmt.Errorf("%s, %w", op, ErrInvalidCredentials)
		}

		log.Error("Password check failed", sl.Err(err))
		return "", fmt.Errorf("%s, %w", op, err)
	}
	
	jwtToken, err := s.jwtService.NewToken(existingUser)
	
	if err != nil {
		log.Error("Failed to produce JWT", sl.Err(err))
		return "", fmt.Errorf("%s, %w", op, err)
	}
	
	log.Info("Jwt token succesfully created")

	return jwtToken, nil
}

func (s *AuthService) RetryEmailVerification(ctx context.Context, email string) error {
	op := "AuthService.RetryEmailVerification"
	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("Retrying to verify Email")

	existingUser, err := s.userService.FindUserByEmail(ctx, email)
	
	if err != nil {
		log.Error("Failed to check user email", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if existingUser.EmailVerified {
		log.Error("Email already verified", sl.Err(ErrEmailVerified))
		return fmt.Errorf("%s, %w", op, ErrEmailVerified)
	}

	return s.sendVerificationMail(existingUser.Email, existingUser.ID)
}

func (s *AuthService) ActivateUser(verificationToken string) error {
	op := "AuthService.ActivateUser"
	log := s.log.With(
		slog.String("op", op),
	)
	log.Info("Activating user")
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
			log.Error("Failed to get verification token info", sl.Err(res.Error))
			return res.Error
		}
		
		// Если условия поиска токена для обновления не выполнятся,
		// ошибки не будет но в rows affected будет 0
		if res.RowsAffected == 0 {
			log.Error("Verification token is invalid", sl.Err(ErrVerificationTokenInvalid))
			return ErrVerificationTokenInvalid
		}
		
		return s.userService.SetUserEmailVerified(tx, token.UserID, true)
		
	})
	// обновление токена и флага активации пользователя
	// в разных операциях, опять же пользователь может остаться не
	// активированным, а токен протухнет
	// Решается ли это ретраем активации почты?
	
	if tokenUpdateErr != nil {
		log.Error("Failed to update verification token", sl.Err(tokenUpdateErr))
		return fmt.Errorf("AuthService: ActivateUser: %w", tokenUpdateErr)
	}

	// сервис пользователя
	log.Info("User succesfully activated")
	return nil
}

// TODO мб покрыть ActivateUser тестами.
// TODO далее внедрять JWT
func (s *AuthService) Auth(ctx context.Context, email string, password string) (user.User, error) {
	// TODO где-то я подменяю обраение к модели напрямую через обращение к сервису, но не вот тут, например,
	// где мне надо вернуть пустого пользака
	op := "AuthService.Auth"

	// Создаём локальный объект логгера с доп. полями, содержащими полезную инфу
    // о текущем вызове функции
	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	
	log.Info("registering a user")

	_, err := s.userService.FindUserByEmail(ctx, email)

	if err != nil {
		if !errors.Is(err, user.ErrUserNotFound) {
			log.Error("failed to check user email", sl.Err(err))
			
			return user.User{}, fmt.Errorf("%s: %w", op, err)
		}
	}
	
	if err == nil {
		log.Error("email already used", sl.Err(ErrEmailAlreadyUsed))
		return user.User{}, fmt.Errorf("%s: %w", op, ErrEmailAlreadyUsed)
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		s.authConfig.BcryptCost,
	)
	
	if err != nil {
		log.Error("failed to hash password", sl.Err(err))
		return user.User{}, fmt.Errorf("%s: %w",op, err)
	}

	createdUser, createUserErr := s.userService.CreateUser(email, string(hashedPassword))

	if createUserErr != nil {
		var pgErr *pgconn.PgError
		log.Error("failed to save user", sl.Err(err))
		
		if errors.As(createUserErr, &pgErr) && pgErr.Code == "23505" {
			return user.User{}, fmt.Errorf("%s: %w", op, ErrEmailAlreadyUsed)
		}
		
		return user.User{}, fmt.Errorf("%s: %w",op, createUserErr)
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
		log.Error("failed to send verification mail", sl.Err(err))
		return user.User{}, fmt.Errorf("%s: %w",op, err)
	}

	log.Info("User succesfully created")

	return createdUser, nil
}

func (s *AuthService) sendVerificationMail(email string, userID uint) error {
	op := "AuthService.sendVerificationMail"
	// TODO тут и логика сохранения токена и отправки email
	// Возможно надо раскидать по методам, а также обеспечить консистентность. Сейчас я просто удалаю токен из БД при ошибке
	// отправки письма. Но мб тут надо внедрить outbox паттерн

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
		slog.Uint64("userID", uint64(userID)),
	)

	log.Info("Sending verification email")

	verificationToken, err := s.createVerificationToken()

	if err != nil {
		log.Error("Failed to create verification token", sl.Err(err))
		return fmt.Errorf("AuthService: sendVerificationMail: %w", err)
	}
	
	tokenHash := s.hashVerificationToken(verificationToken)
	
	tokenModel := VerificationToken{
		Hash:      tokenHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.authConfig.VerificationTokenTTL),
	}
	
	if err := s.DB.Create(&tokenModel).Error; err != nil {
		log.Error("Failed to save verification token", sl.Err(err))
		return fmt.Errorf("AuthService: sendVerificationMail: createToken: %w", err)
	}
	
	if err := s.mailQueue.AddVerificationJob(email, verificationToken); err != nil {
		log.Error("Failed to add verification job", sl.Err(err))

		if delErr := s.DB.Delete(&tokenModel).Error; delErr != nil {
			log.Error("Failed to rollback verification token, after error", sl.Err(delErr))
		}

		return fmt.Errorf("AuthService: sendVerificationMail: send mail: %w", err)
	}

	return nil
}

func (s *AuthService) hashVerificationToken(token string) string {
	op := "AuthService.hashVerificationToken"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("Hashing verification token")

	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}

func (s *AuthService) createVerificationToken() (string, error) {
	op := "AuthService.createVerificationToken"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("Creating verification token")

	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		log.Error("Failed to create verification token", sl.Err(err))
		return "", fmt.Errorf("AuthService: createVerificationToken: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *AuthService) comparePassword(hashedPassword []byte, password []byte) error {
	op := "AuthService.comparePassword"
	log := s.log.With(
		slog.String("op", op),
	)
	log.Info("Comparing password with hash")
	
	if err := bcrypt.CompareHashAndPassword(hashedPassword, password); err != nil {
		log.Error("Password comparission failed", sl.Err(err))
		return fmt.Errorf("%s, %w", op, err)
	}

	return nil
}