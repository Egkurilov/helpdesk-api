package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

var logger = logrus.New()

func JWTMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		logger.Infof("Received Authorization header: %s", authHeader)
		if authHeader == "" {
			logger.Warn("No Authorization header provided")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется токен авторизации"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warnf("Invalid Authorization header format: %s", authHeader)
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
			logger.Warnf("Token parsing failed or invalid: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный или истекший токен"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Warn("Failed to extract claims from token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные токена"})
			return
		}

		role, _ := claims["role"].(string)
		logger.Infof("Extracted role: %s", role)
		if role == "" {
			logger.Warn("Role not found in token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Роль не указана в токене"})
			return
		}

		if role == "user" {
			telegramID, ok := claims["telegram_id"].(string)
			if !ok || telegramID == "" {
				logger.Warn("No telegram_id in user token")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Нет telegram_id в токене"})
				return
			}
			c.Set("telegram_id", telegramID)
		} else if role == "operator" {
			username, ok := claims["username"].(string)
			if !ok || username == "" {
				logger.Warn("No username in operator token")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Нет username в токене"})
				return
			}
			c.Set("username", username)
		}

		c.Set("role", role)
		c.Set("claims", claims) // Устанавливаем claims для operatorMiddleware
		logger.Info("JWT middleware passed successfully")
		c.Next()
	}
}
