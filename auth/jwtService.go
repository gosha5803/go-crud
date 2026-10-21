package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/user"
)

type JwtService struct {
	authConfig *initializers.AuthConfig
}

func NewJwtSevice(
	authConfig *initializers.AuthConfig,
) *JwtService {
	return &JwtService{authConfig: authConfig}
}

func (s *JwtService) NewToken(user user.User) (string, error) {
	op := "JwtService.NewToken"
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["userId"] = user.ID
	claims["email"] = user.Email
	// Время жизни
	claims["exp"] = time.Now().Add(s.authConfig.JwtTokenTTL).Unix()

	tokenString, err := token.SignedString([]byte(s.authConfig.JwtTokenSecret))

	if err != nil {
		return "", fmt.Errorf("%s, %w", op, err)
	}

	return tokenString, nil
}