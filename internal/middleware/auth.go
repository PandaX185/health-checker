package middleware

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !parsedToken.Valid {
			if errors.Is(err, jwt.ErrTokenExpired) {
				ctx.AbortWithStatusJSON(401, gin.H{"error": "Token expired"})
				return
			}
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Invalid token claims"})
			return
		}

		newContext := context.WithValue(ctx.Request.Context(), "user_id", claims["user_id"])

		ctx.Request = ctx.Request.WithContext(newContext)
		ctx.Next()
	}
}
