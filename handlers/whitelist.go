package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"helpdesk-api/models"
	"log"
	"net/http"
	"time"
)

type StandConfig struct {
	Stands map[string]string
}

var standEndpoints StandConfig

// LoadEndpoints загружает endpoints из базы данных
func LoadEndpoints(db *gorm.DB) {
	var endpoints []models.Endpoint
	if err := db.Find(&endpoints).Error; err != nil {
		log.Fatalf("Failed to load endpoints from database: %v", err)
	}

	standEndpoints.Stands = make(map[string]string)
	for _, endpoint := range endpoints {
		standEndpoints.Stands[endpoint.Name] = endpoint.URL
	}
	log.Println("Loaded stand endpoints:", standEndpoints)
}

// Пример использования в обработчике
func SomeHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Доступ к endpoints
		for name, url := range standEndpoints.Stands {
			c.JSON(200, gin.H{"name": name, "url": url})
		}
	}
}

type WhitelistRequestInput struct {
	Text   string `json:"text" binding:"required"`
	ChatID int64  `json:"chatId" binding:"required"`
	From   string `json:"from" binding:"required,oneof=dev ift psi prom"`
	User   struct {
		ID           int64  `json:"id" binding:"required"`
		IsBot        bool   `json:"is_bot"`
		FirstName    string `json:"first_name" binding:"required"`
		LastName     string `json:"last_name" binding:"required"`
		Username     string `json:"username" binding:"required"`
		LanguageCode string `json:"language_code"`
	} `json:"user" binding:"required"`
}

type WhitelistEditInput struct {
	Permission string `json:"permission" binding:"required,oneof=approve deny"` // "approve" или "deny"
}

// AddWhitelistRequest godoc
// @Summary Создать новую заявку в whitelist
// @Description Добавляет новую заявку в whitelist со статусом "pending". Если заявка уже существует (по telegram_id и from), возвращает 200 OK.
// @Tags whitelist
// @Accept json
// @Produce json
// @Param request body WhitelistRequestInput true "Данные заявки"
// @Success 201 {object} map[string]interface{} "message: Запрос создан, id: <id>"
// @Success 200 {object} map[string]string "message: Запрос уже существует"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Security BearerAuth
// @Router /whitelist [post]
func AddWhitelistRequest(c *gin.Context, db *gorm.DB) {
	var input WhitelistRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	whitelist := models.Whitelist{
		TelegramID:   fmt.Sprintf("%d", input.User.ID),
		Text:         input.Text,
		ChatID:       input.ChatID,
		From:         input.From,
		FirstName:    input.User.FirstName,
		LastName:     input.User.LastName,
		Username:     input.User.Username,
		LanguageCode: input.User.LanguageCode,
		Permission:   "pending", // Дефолтное значение
	}

	// Попытка создать запись
	result := db.Create(&whitelist)
	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusOK, gin.H{"message": "Запрос уже существует"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать запись в whitelist: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Запрос создан",
		"id":      whitelist.ID,
	})
}

// EditWhitelist godoc
// @Summary Изменить статус заявки в whitelist
// @Description Обновляет статус заявки в whitelist ("approve" или "deny") и уведомляет стенд при одобрении
// @Tags whitelist
// @Accept json
// @Produce json
// @Param telegram_id path string true "Telegram ID пользователя"
// @Param request body WhitelistEditInput true "Новое значение permission"
// @Success 200 {object} map[string]string "message: OK"
// @Failure 400 {object} map[string]string "error"
// @Failure 404 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Security BearerAuth
// @Router /operator/whitelist/{telegram_id}/edit [post]
func EditWhitelist(c *gin.Context, db *gorm.DB) {
	telegramID := c.Param("telegram_id")
	var input WhitelistEditInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var whitelist models.Whitelist
	if err := db.Where("telegram_id = ?", telegramID).First(&whitelist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден в whitelist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки whitelist: " + err.Error()})
		}
		return
	}

	// Обновляем статус
	whitelist.Permission = input.Permission
	if err := db.Save(&whitelist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить доступ: " + err.Error()})
		return
	}

	// Если статус "approve", уведомляем стенд
	if input.Permission == "approve" {
		standEndpoint, exists := standEndpoints.Stands[whitelist.From]
		if !exists {
			log.Printf("No endpoint found for stand: %s", whitelist.From)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Неизвестный стенд: %s", whitelist.From)})
			return
		}

		// Формируем новый payload согласно требованиям
		payload := map[string]interface{}{
			"chatId":  whitelist.ChatID,
			"message": "Тестовое сообщение2",
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal payload for stand %s: %v", whitelist.From, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare request to stand"})
			return
		}

		// Отправляем запрос с тайм-аутом
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(standEndpoint, "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Printf("Failed to notify stand %s at %s: %v", whitelist.From, standEndpoint, err)
			// Не прерываем выполнение, просто логируем ошибку
		} else if resp.StatusCode != http.StatusOK {
			log.Printf("Stand %s at %s returned status %d", whitelist.From, standEndpoint, resp.StatusCode)
		} else {
			log.Printf("Successfully notified stand %s at %s", whitelist.From, standEndpoint)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

// GetWhitelistPending godoc
// @Summary Получить список ожидающих заявок whitelist
// @Description Возвращает все записи whitelist со статусом "pending"
// @Tags whitelist
// @Produce json
// @Success 200 {array} models.Whitelist
// @Failure 500 {object} map[string]string "error"
// @Security BearerAuth
// @Router /operator/whitelist [get]
func GetWhitelistPending(c *gin.Context, db *gorm.DB) {
	var pendingUsers []models.Whitelist
	if err := db.Where("permission = ?", "pending").Find(&pendingUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список ожидания: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, pendingUsers)
}

// GetWhitelistAll godoc
// @Summary Получить все записи whitelist
// @Description Возвращает все записи whitelist независимо от статуса
// @Tags whitelist
// @Produce json
// @Success 200 {array} models.Whitelist
// @Failure 500 {object} map[string]string "error"
// @Security BearerAuth
// @Router /operator/whitelist/all [get]
func GetWhitelistAll(c *gin.Context, db *gorm.DB) {
	var allUsers []models.Whitelist
	if err := db.Find(&allUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список whitelist: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, allUsers)
}
