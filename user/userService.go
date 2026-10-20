package user

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type IMailService interface {
	SendVerificationMail(string, string)
}

type UserService struct {
	DB *gorm.DB
}

func NewUserService(DB *gorm.DB) *UserService {
	return &UserService{DB: DB}
}

func (s *UserService) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	op := "UserService.FindUserByEmail"

	result := s.DB.
		WithContext(ctx).
		Where("email = ?", email).
		First(&user)
		
	if result.Error!= nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}

		return User{}, fmt.Errorf("%s: %w", op, result.Error)
	}

	return user, nil
}

func (s *UserService) CreateUser(email string, password string) (User, error) {
	user := User{
		Password: password,
		Email:    email,
	}

	if err := s.DB.Create(&user).Error; err != nil {
		return User{}, fmt.Errorf("UserService: createUser: %w", err)
	}

	return user, nil
}

func (s *UserService) SetUserEmailVerified(tx *gorm.DB, userID uint, isVerified bool) error {

	if userUpdateErr := tx.Model(&User{}).
		Where("id = ?", userID).
		Update("email_verified", isVerified).
		Error; userUpdateErr != nil {
		return fmt.Errorf("UserService: setUserEmailVerified: %w", userUpdateErr)
	}

	return nil
}
