package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB        *gorm.DB
	JWTSecret string // ✅ просто объявляем, но не инициализируем
)

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("No .env file found")
	}

	JWTSecret = os.Getenv("JWT_SECRET") // ⬅️ теперь это сработает!
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET is not set!")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	log.Println("Database connected successfully!")
}
