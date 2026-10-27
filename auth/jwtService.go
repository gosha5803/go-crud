package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gosha5803/go-crud/initializers"
	"github.com/gosha5803/go-crud/user"
)

type JwtClaims struct {
	jwt.RegisteredClaims
}

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
	now := time.Now()
	claims := JwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.authConfig.JwtTokenTTL)),
			Subject: fmt.Sprint(user.ID),
			IssuedAt: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.authConfig.JwtTokenSecret))

	if err != nil {
		return "", fmt.Errorf("%s, %w", op, err)
	}

	return tokenString, nil
}