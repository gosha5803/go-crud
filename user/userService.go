package user

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(DB *gorm.DB) *UserService {
	return &UserService{DB: DB}
}

func (s *UserService) Auth(dto AuthReqDto) (User, error) {
	var existingCount int64

	// validate that email is Unique
	if err := s.DB.Model(&User{}).
		Where("email = ?", dto.Email).
		Count(&existingCount).Error; err != nil {

		return User{}, fmt.Errorf("userService: Login: %w", err)
	}

	if existingCount > 0 {
		// TODO расширить ошибку контроллера
		return User{}, fmt.Errorf("userService: Login: %w", ErrEmailAlreadyUsed)
	}

	// TODO вынести
	salt := 10
	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), salt)

	if err != nil {
		return User{}, fmt.Errorf("userService: Login: %w", err)
	}

	user := User{
		Password: string(hashedPassword),
		Email:    dto.Email,
	}

	result := s.DB.Create(&user)

	if result.Error != nil {
		// TODO вынести
		return User{}, fmt.Errorf("userService: Login: %w", err)
	}

	return user, nil
}
