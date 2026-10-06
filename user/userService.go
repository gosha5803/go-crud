package user

import (
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

func (s *UserService) IsEmailExist(email string) (bool, error) {
	var existingCount int64

	if err := s.DB.Model(&User{}).
		Where("email = ?", email).
		Count(&existingCount).Error; err != nil {

		return true, fmt.Errorf("userService: Login: %w", err)
	}

	if existingCount > 0 {
		return true, nil
	}

	return false, nil
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
