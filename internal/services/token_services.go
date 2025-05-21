package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"secure-messenger/internal/models"
	"time"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateRefreshToken(db *gorm.DB, userID uint) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	refreshToken := hex.EncodeToString(tokenBytes)

	rt := models.RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := db.Create(&rt).Error; err != nil {
		return "", err
	}
	return refreshToken, nil
}

func ValidateRefreshToken(db *gorm.DB, token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	if err := db.Where("token = ?", token).First(&rt).Error; err != nil {
		return nil, err
	}
	if rt.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired")
	}
	return &rt, nil
}
