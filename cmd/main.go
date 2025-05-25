package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"secure-messenger/config"
	"secure-messenger/internal/handlers"
	"secure-messenger/internal/models"
	"secure-messenger/internal/repository"
	"secure-messenger/internal/services"
)

func main() {
	config.InitDB()

	// ✅ Автоматическая миграция таблиц
	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Message{},
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	r := gin.Default()

	// ===== API Group =====
	api := r.Group("/api")

	// --- Auth Endpoints ---
	api.POST("/register", handlers.RegisterWithDB(config.DB))
	api.POST("/login", handlers.Login)
	api.POST("/refresh", handlers.Refresh)

	// --- Protected routes ---
	api.GET("/profile", handlers.AuthMiddleware(""), handlers.ProfileHandler(config.DB))

	// --- Admin routes ---
	admin := api.Group("/admin")
	admin.Use(handlers.AuthMiddleware("admin"))
	{
		admin.GET("/users", handlers.GetAllUsersWithDB(config.DB))
		admin.DELETE("/users/:id", handlers.DeleteUserWithDB(config.DB))
		admin.PUT("/users/:id", handlers.UpdateUserWithDB(config.DB))
	}

	// ===== Messaging Dependencies =====
	messageRepo := repository.NewMessageRepository(config.DB)
	messageService := services.NewMessageService(messageRepo, config.AESSecretKey) // ✅ передаём ключ
	messageHandler := handlers.NewMessageHandler(messageService)

	// --- Messaging Endpoints ---
	api.Use(handlers.AuthMiddleware("")) // 🔐 Require auth for message routes
	{
		api.POST("/messages/send", messageHandler.SendMessage)
		api.GET("/messages", messageHandler.GetMessages)
		api.DELETE("/messages/:id", messageHandler.DeleteMessage)
	}

	// Запуск
	fmt.Println("Server running at :8081")
	_ = r.Run(":8081")
}
