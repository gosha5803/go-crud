package auth

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret []byte, log *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
			log.Warn("missing or malformed authorization header",
				slog.String("path", ctx.Request.URL.Path),
				slog.String("ip", ctx.ClientIP()),
			)
			// Просто ctx.JSON возвращал управление дальше
			ctx.AbortWithStatusJSON(ErrUnauthorized.Status, ErrUnauthorized)
			return 
		}
		
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				 log.Warn("unexpected jwt signing method",
					slog.String("alg", fmt.Sprint(t.Header["alg"])),
					slog.String("ip", ctx.ClientIP()),
				)
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"]) 
			}
			
			// Возвращаемый секрет используется функцие парсинга 
			// чтобы распарсить токен
			return secret, nil
			}, jwt.WithValidMethods([]string{"HS256"}))
			
		if err != nil || !token.Valid {
			log.Warn("invalid jwt token",
				slog.String("path", ctx.Request.URL.Path),
				slog.String("ip", ctx.ClientIP()),
				slog.Any("err", err), // "token expired", "signature invalid", ...
			)
			ctx.AbortWithStatusJSON(ErrUnauthorized.Status, ErrUnauthorized)
			return 
		}

		// TODO Пройтись по данным, если надо

		ctx.Next()
	}
}