package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"helpdesk-api/models"
	"net/http"
)

type WhitelistRequestInput struct {
	Text   string `json:"text" binding:"required"`
	ChatID int64  `json:"chatId" binding:"required"`
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
	Perm *bool `json:"perm" binding:"required"`
}

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
		FirstName:    input.User.FirstName,
		LastName:     input.User.LastName,
		Username:     input.User.Username,
		LanguageCode: input.User.LanguageCode,
		Permission:   0, // Pending
	}

	if err := db.Create(&whitelist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать запись в whitelist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Запрос получен"})
}

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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки whitelist"})
		}
		return
	}

	if input.Perm != nil && *input.Perm {
		whitelist.Permission = 1 // Одобрено
	} else {
		whitelist.Permission = 2 // Отклонено
	}
	if err := db.Save(&whitelist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить доступ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func GetWhitelistPending(c *gin.Context, db *gorm.DB) {
	var pendingUsers []models.Whitelist
	if err := db.Where("permission = ?", 0).Find(&pendingUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список ожидания"})
		return
	}
	c.JSON(http.StatusOK, pendingUsers)
}

func GetWhitelistAll(c *gin.Context, db *gorm.DB) {
	var allUsers []models.Whitelist
	if err := db.Find(&allUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список whitelist"})
		return
	}
	c.JSON(http.StatusOK, allUsers)
}
