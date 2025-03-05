package handlers

import (
	"fmt"
	"net/http"
	"time"

	"helpdesk-api/config"
	"helpdesk-api/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TokenInput структура для входных данных получения токена пользователя
type TokenInput struct {
	TelegramID string `json:"telegram_id" binding:"required" example:"88376478"`
}

// OperatorLoginInput структура для входных данных логина оператора
type OperatorLoginInput struct {
	Username string `json:"username" binding:"required" example:"operator1"`
	Password string `json:"password" binding:"required" example:"securepassword"`
}

// RegisterConsumer godoc
// @Summary Получить JWT-токен для пользователя
// @Description Регистрирует или возвращает токен для пользователя по Telegram ID
// @Tags auth
// @Accept json
// @Produce json
// @Param input body TokenInput true "Telegram ID пользователя"
// @Success 200 {object} map[string]string "access: JWT-токен"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /consumers/token/ [post]
func RegisterConsumer(c *gin.Context, db *gorm.DB, cfg *config.Config) {
	var input TokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("telegram_id = ?", input.TelegramID).First(&user).Error; err != nil {
		user = models.User{
			TelegramID: input.TelegramID,
		}
		db.Create(&user)
	}

	token, err := generateJWT(user, "user", cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать токен"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access": token})
}

// Login godoc
// @Summary Логин оператора
// @Description Авторизует оператора по логину и паролю, возвращает JWT-токен
// @Tags auth
// @Accept json
// @Produce json
// @Param input body OperatorLoginInput true "Данные оператора"
// @Success 200 {object} map[string]string "access: JWT-токен"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /token/ [post]
func Login(c *gin.Context, db *gorm.DB, cfg *config.Config) {
	var input OperatorLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var operator models.Operator
	if err := db.Where("username = ?", input.Username).First(&operator).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(operator.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	token, err := generateJWT(operator, operator.Role, cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать токен"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access": token})
}

// Обновленный generateJWT для поддержки ролей
func generateJWT(entity interface{}, role string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	}

	switch e := entity.(type) {
	case models.User:
		claims["telegram_id"] = e.TelegramID
	case models.Operator:
		claims["username"] = e.Username
	default:
		return "", fmt.Errorf("неизвестный тип сущности")
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// Logout godoc
// @Summary Выход оператора
// @Description Подтверждает выход оператора; клиент должен удалить токен
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "message: Успешный выход"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Router /logout/ [post]
func Logout(c *gin.Context, db *gorm.DB, cfg *config.Config) {
	role, exists := c.Get("role")
	if !exists || role != "operator" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Доступ только для операторов"})
		return
	}

	// Здесь можно добавить дополнительную логику, если нужно
	// Например, логирование выхода оператора
	username, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{"message": "Успешный выход", "username": username})
}

// CloseTicketOperator — закрытие тикета (только для операторов)
func CloseTicketOperator(c *gin.Context, db *gorm.DB) {
	ticketID := c.Param("ticket_id")

	var ticket models.Ticket
	if err := db.Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	if ticket.Status == "CLOSED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket already closed"})
		return
	}

	ticket.Status = "CLOSED"
	if err := db.Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close ticket"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Ticket closed successfully"})
}
