package handlers

import (
	"net/http"
	"time"

	"helpdesk-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// createTicketInput структура для входных данных создания тикета
type createTicketInput struct {
	Subject     string `json:"subject" binding:"required" example:"Проблема с продуктом"`     // Тема тикета
	Description string `json:"description" binding:"required" example:"Описание проблемы..."` // Описание проблемы
	Source      string `json:"source" binding:"required" example:"Telegram"`                  // Источник тикета
}

// CreateTicket godoc
// @Summary Создать новый тикет
// @Description Создает тикет от текущего пользователя
// @Tags tickets
// @Accept json
// @Produce json
// @Param ticket body createTicketInput true "Данные тикета"
// @Success 201 {object} models.Ticket
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Security BearerAuth
// @Router /tickets/create [post]
func CreateTicket(c *gin.Context, db *gorm.DB) {
	var input createTicketInput // Используем именованную структуру
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	telegramIDVal, exists := c.Get("telegram_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	telegramID, ok := telegramIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid telegram_id type"})
		return
	}

	var user models.User
	if err := db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	ticket := models.Ticket{
		UserID:      user.ID,
		Subject:     input.Subject,
		Description: input.Description,
		Source:      input.Source,
		Status:      "OPEN",
	}
	if err := db.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

// ListTickets godoc
// @Summary Получить список тикетов
// @Description Возвращает все тикеты для оператора или тикеты текущего пользователя
// @Tags tickets
// @Produce json
// @Success 200 {array} models.Ticket
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Security BearerAuth
// @Router /tickets/ [get]
func ListTickets(c *gin.Context, db *gorm.DB) {
	role, _ := c.Get("role")
	if role == "operator" {
		var tickets []models.Ticket
		if err := db.Find(&tickets).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching tickets"})
			return
		}
		c.JSON(http.StatusOK, tickets)
		return
	}

	// Логика для пользователей
	telegramIDInterface, exists := c.Get("telegram_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	telegramID, ok := telegramIDInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid telegram_id in token"})
		return
	}

	var user models.User
	if err := db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var tickets []models.Ticket
	if err := db.Where("user_id = ?", user.ID).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

// addMessageInput структура для входных данных сообщения
type addMessageInput struct {
	Sender    string `json:"sender" binding:"required" example:"user"`
	Recipient string `json:"recipient" binding:"required" example:"operator"`
	Content   string `json:"content" binding:"required" example:"Сообщение"`
}

// AddMessage godoc
// @Summary Добавить сообщение в тикет
// @Description Добавляет сообщение в указанный тикет
// @Tags messages
// @Accept json
// @Produce json
// @Param ticket_id path string true "ID тикета"
// @Param message body addMessageInput true "Данные сообщения"
// @Success 201 {object} models.Message
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Security BearerAuth
// @Router /tickets/{ticket_id}/messages/ [post]
func AddMessage(c *gin.Context, db *gorm.DB) {
	role, _ := c.Get("role")
	ticketID := c.Param("ticket_id")
	var ticket models.Ticket
	if err := db.Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	var input addMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Sender != "user" && input.Sender != "operator" || input.Recipient != "user" && input.Recipient != "operator" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sender and Recipient must be 'user' or 'operator'"})
		return
	}

	if role == "operator" {
		if input.Sender != "operator" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Operators can only send as 'operator'"})
			return
		}
	} else {
		telegramIDVal, exists := c.Get("telegram_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		telegramID, ok := telegramIDVal.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid telegram_id type"})
			return
		}

		var user models.User
		if err := db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		if input.Sender == "user" && ticket.UserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the ticket owner can send as 'user'"})
			return
		}
		if input.Sender == "operator" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only operators can send as 'operator'"})
			return
		}
	}

	message := models.Message{
		TicketID:  ticket.ID,
		Sender:    input.Sender,
		Recipient: input.Recipient,
		Content:   input.Content,
	}
	if err := db.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetTicketHistory godoc
// @Summary Получить историю сообщений тикета
// @Description Возвращает все сообщения для указанного тикета
// @Tags messages
// @Produce json
// @Param ticket_id path string true "ID тикета"
// @Success 200 {array} models.Message
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Security BearerAuth
// @Router /tickets/{ticket_id}/messages/ [get]
func GetTicketHistory(c *gin.Context, db *gorm.DB) {
	role, _ := c.Get("role")
	ticketID := c.Param("ticket_id")
	var ticket models.Ticket
	if err := db.Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	if role != "operator" {
		telegramIDVal, exists := c.Get("telegram_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		telegramID, ok := telegramIDVal.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid telegram_id type"})
			return
		}

		var user models.User
		if err := db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		if ticket.UserID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only view your own tickets"})
			return
		}
	}

	var messages []models.Message
	if err := db.Where("ticket_id = ?", ticket.ID).Order("timestamp asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching messages"})
		return
	}
	c.JSON(http.StatusOK, messages)
}

// CloseTicket godoc
// @Summary Закрыть тикет
// @Description Закрывает указанный тикет
// @Tags tickets
// @Produce json
// @Param ticket_id path string true "ID тикета"
// @Success 200 {object} models.Ticket
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Security BearerAuth
// @Router /tickets/{ticket_id}/close/ [post]
func CloseTicket(c *gin.Context, db *gorm.DB) {
	ticketID := c.Param("ticket_id")

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var ticket models.Ticket
	var err error

	switch role {
	case "user":
		telegramID, _ := c.Get("telegram_id")
		err = db.Where("id = ? AND user_id = (SELECT id FROM users WHERE telegram_id = ?)", ticketID, telegramID).First(&ticket).Error
	case "operator":
		// Оператор может закрывать любой тикет
		err = db.Where("id = ?", ticketID).First(&ticket).Error
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
		return
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	if ticket.Status == "CLOSED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket already closed"})
		return
	}

	ticket.Status = "CLOSED"
	ticket.ClosedBy = role.(string) // Сохраняем, кто закрыл тикет
	ticket.ClosedAt = time.Now()
	if err := db.Save(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket closed successfully", "ticket": ticket})
}
