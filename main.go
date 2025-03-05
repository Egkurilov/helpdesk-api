package main

import (
	"github.com/gin-contrib/cors"
	"golang.org/x/crypto/bcrypt"
	"helpdesk-api/config"
	"helpdesk-api/models"
	"helpdesk-api/routes"
	"helpdesk-api/utils"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	_ "helpdesk-api/docs"
)

// @title Helpdesk API
// @version 1.0
// @description API для системы поддержки пользователей с тикетами и перепиской
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.LoadConfig()
	logger := utils.InitLogger()

	dsn := "host=" + cfg.DBHost + " user=" + cfg.DBUser + " password=" + cfg.DBPassword + " dbname=" + cfg.DBName + " port=" + cfg.DBPort + " sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Ошибка подключения к базе данных: ", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Ticket{}, &models.Message{}, &models.Operator{})
	if err != nil {
		logger.Fatal("Ошибка миграции: ", err)
	}

	// Создание тестового оператора
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("securepassword"), bcrypt.DefaultCost)
	testOperator := models.Operator{
		Username: "operator1",
		Password: string(hashedPassword),
		Role:     "operator",
	}
	db.FirstOrCreate(&testOperator, models.Operator{Username: "operator1"})

	router := gin.Default()

	// Настройка CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8001", "http://localhost:8000"}, // Разрешаем фронтенд
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.SetupRoutes(router, db, cfg, logger) // Здесь добавится новый маршрут для оператора

	// Swagger
	router.GET("/swagger/*any", func(c *gin.Context) {
		logger.Info("Serving Swagger request: ", c.Request.URL.Path)
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		if c.Writer.Status() >= 400 {
			logger.Error("Failed to serve Swagger: ", c.Writer.Status())
		}
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
