package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"helpdesk-api/config"
	"helpdesk-api/handlers"
	"helpdesk-api/middleware"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config, logger *logrus.Logger) {
	public := router.Group("/api")
	{
		public.POST("/consumers/token/", func(c *gin.Context) {
			handlers.RegisterConsumer(c, db, cfg)
		})
		public.POST("/token/", func(c *gin.Context) {
			handlers.Login(c, db, cfg)
		})
	}

	protected := router.Group("/api")
	protected.Use(middleware.JWTMiddleware(cfg.JWTSecret))
	{
		// Общие маршруты
		protected.POST("/tickets/create", func(c *gin.Context) {
			handlers.CreateTicket(c, db)
		})
		protected.GET("/tickets/", func(c *gin.Context) {
			handlers.ListTickets(c, db)
		})
		protected.POST("/tickets/:ticket_id/messages/", func(c *gin.Context) {
			handlers.AddMessage(c, db)
		})
		protected.GET("/tickets/:ticket_id/messages/", func(c *gin.Context) {
			handlers.GetTicketHistory(c, db)
		})
		protected.POST("/tickets/:ticket_id/close/", func(c *gin.Context) {
			handlers.CloseTicket(c, db)
		})
		protected.POST("/logout/", func(c *gin.Context) {
			handlers.Logout(c, db, cfg)
		})

		// Маршруты только для операторов
		operator := protected.Group("/operator")
		operator.Use(operatorMiddleware())
		{
			operator.POST("/ticket/:ticket_id/close/", func(c *gin.Context) {
				handlers.CloseTicketOperator(c, db)
			})
		}
	}
}

// operatorMiddleware проверяет роль "operator"
func operatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		role := claims.(map[string]interface{})["role"].(string)
		if role != "operator" {
			c.JSON(403, gin.H{"error": "Forbidden: Operator access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
