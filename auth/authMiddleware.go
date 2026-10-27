package auth

import (
	"log/slog"
	"strconv"
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
		
		token, err := jwt.ParseWithClaims(
			tokenString, 
			&JwtClaims{},
			func(t *jwt.Token) (any, error) {
				return secret, nil
			}, 
			jwt.WithValidMethods([]string{"HS256"}),
		)
			
		if err != nil || !token.Valid {
			log.Warn("invalid jwt token",
				slog.String("path", ctx.Request.URL.Path),
				slog.String("ip", ctx.ClientIP()),
				slog.Any("err", err), // "token expired", "signature invalid", ...
			)
			ctx.AbortWithStatusJSON(ErrUnauthorized.Status, ErrUnauthorized)
			return 
		}

		claims, ok := token.Claims.(*JwtClaims)
		
		if !ok {
			log.Warn("missing jwt claims",
				slog.String("path", ctx.Request.URL.Path),
				slog.String("ip", ctx.ClientIP()),
			)
			ctx.AbortWithStatusJSON(ErrUnauthorized.Status, ErrUnauthorized)
			return 
		}

		userId, err := strconv.ParseUint(claims.Subject, 10, 0) 
	
		if err != nil {
			log.Warn("failed to convert jwt sub",
				slog.String("path", ctx.Request.URL.Path),
				slog.String("ip", ctx.ClientIP()),
			)
			ctx.AbortWithStatusJSON(ErrUnauthorized.Status, ErrUnauthorized)
			return 
		}

		ctx.Set(UserIdKey, uint(userId))
		ctx.Next()
	}
}