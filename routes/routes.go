package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"helpdesk-api/config"
	"helpdesk-api/handlers"
	"helpdesk-api/middleware"
	"helpdesk-api/models"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config, logger *logrus.Logger) {
	handlers.LoadEndpoints(db)

	public := router.Group("/api")
	{
		public.POST("/consumers/token/", func(c *gin.Context) {
			handlers.RegisterConsumer(c, db, cfg)
		})
		public.POST("/token/", func(c *gin.Context) {
			handlers.Login(c, db, cfg)
		})
		public.POST("/whitelist", func(c *gin.Context) {
			handlers.AddWhitelistRequest(c, db)
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

			// Маршруты для управления настройками
			operator.GET("/settings/", func(c *gin.Context) {
				var endpoints []models.Endpoint
				if err := db.Find(&endpoints).Error; err != nil {
					logger.Errorf("Failed to fetch endpoints: %v", err)
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				logger.Infof("Fetched %d endpoints", len(endpoints))
				c.JSON(200, endpoints)
			})

			operator.POST("/settings/", func(c *gin.Context) {
				var endpoint models.Endpoint
				if err := c.ShouldBindJSON(&endpoint); err != nil {
					logger.Errorf("Invalid JSON for endpoint: %v", err)
					c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
					return
				}
				if endpoint.Name == "" || endpoint.URL == "" {
					logger.Errorf("Missing required fields: %+v", endpoint)
					c.JSON(400, gin.H{"error": "Name and URL are required"})
					return
				}
				if err := db.Create(&endpoint).Error; err != nil {
					logger.Errorf("Failed to create endpoint: %v", err)
					c.JSON(500, gin.H{"error": "Failed to save endpoint: " + err.Error()})
					return
				}
				handlers.LoadEndpoints(db)
				logger.Infof("Created endpoint: %+v", endpoint)
				c.JSON(201, endpoint)
			})

			operator.PUT("/settings/:id", func(c *gin.Context) {
				id := c.Param("id")
				var endpoint models.Endpoint
				if err := db.First(&endpoint, id).Error; err != nil {
					logger.Errorf("Endpoint not found: %v", err)
					c.JSON(404, gin.H{"error": "Endpoint not found"})
					return
				}
				if err := c.ShouldBindJSON(&endpoint); err != nil {
					logger.Errorf("Invalid JSON for endpoint update: %v", err)
					c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
					return
				}
				if endpoint.Name == "" || endpoint.URL == "" {
					logger.Errorf("Missing required fields: %+v", endpoint)
					c.JSON(400, gin.H{"error": "Name and URL are required"})
					return
				}
				if err := db.Save(&endpoint).Error; err != nil {
					logger.Errorf("Failed to update endpoint: %v", err)
					c.JSON(500, gin.H{"error": "Failed to update endpoint: " + err.Error()})
					return
				}
				handlers.LoadEndpoints(db)
				logger.Infof("Updated endpoint: %+v", endpoint)
				c.JSON(200, endpoint)
			})

			operator.DELETE("/settings/:id", func(c *gin.Context) {
				id := c.Param("id")
				if id == "" {
					logger.Errorf("Invalid endpoint ID: %s", id)
					c.JSON(400, gin.H{"error": "Invalid endpoint ID"})
					return
				}
				if err := db.Delete(&models.Endpoint{}, id).Error; err != nil {
					logger.Errorf("Failed to delete endpoint: %v", err)
					c.JSON(500, gin.H{"error": "Failed to delete endpoint: " + err.Error()})
					return
				}
				handlers.LoadEndpoints(db)
				logger.Infof("Deleted endpoint with ID: %s", id)
				c.JSON(204, nil)
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
