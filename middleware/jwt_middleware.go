package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func JWTMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен авторизации"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный или истекший токен"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные токена"})
			return
		}

		role, _ := claims["role"].(string)
		if role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Роль не указана в токене"})
			return
		}

		if role == "user" {
			telegramID, ok := claims["telegram_id"].(string)
			if !ok || telegramID == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Нет telegram_id в токене"})
				return
			}
			c.Set("telegram_id", telegramID)
		} else if role == "operator" {
			username, ok := claims["username"].(string)
			if !ok || username == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Нет username в токене"})
				return
			}
			c.Set("username", username)
		}

		c.Set("role", role)
		c.Next()
	}
}
