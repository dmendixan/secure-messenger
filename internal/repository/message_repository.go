package repository

import (
	"gorm.io/gorm"
	"secure-messenger/internal/models"
)

type MessageRepository struct {
	DB *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{DB: db}
}

func (r *MessageRepository) CreateMessage(msg *models.Message) error {
	return r.DB.Create(msg).Error
}

func (r *MessageRepository) GetMessagesForUser(userID uint) ([]models.Message, error) {
	var messages []models.Message
	err := r.DB.Where("receiver_id = ?", userID).Or("sender_id = ?", userID).Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) DeleteMessage(id uint, userID uint) error {
	return r.DB.Where("id = ? AND sender_id = ?", id, userID).Delete(&models.Message{}).Error
}
