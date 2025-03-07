package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
		public.POST("/whitelist", func(c *gin.Context) {
			handlers.AddWhitelistRequest(c, db)
		})
		operator := protected.Group("/operator")
		operator.Use(operatorMiddleware())
		{
			operator.POST("/ticket/:ticket_id/close/", func(c *gin.Context) {
				handlers.CloseTicketOperator(c, db)
			})
			operator.POST("/whitelist/:telegram_id/edit", func(c *gin.Context) {
				handlers.EditWhitelist(c, db)
			})
			operator.GET("/whitelist/", func(c *gin.Context) {
				handlers.GetWhitelistPending(c, db)
			})
			operator.GET("/whitelist/all", func(c *gin.Context) {
				handlers.GetWhitelistAll(c, db)
			})
		}
	}
}

func operatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		jwtClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(500, gin.H{"error": "Invalid claims type"})
			c.Abort()
			return
		}

		role, ok := jwtClaims["role"].(string)
		if !ok || role != "operator" {
			c.JSON(403, gin.H{"error": "Forbidden: Operator access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
