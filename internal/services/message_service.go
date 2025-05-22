package services

import (
	"secure-messenger/internal/models"
	"secure-messenger/internal/repository"
	"secure-messenger/pkg/encryption"
)

type MessageService struct {
	Repo *repository.MessageRepository
}

func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{Repo: repo}
}

func (s *MessageService) SendMessage(senderID, receiverID uint, plainText string, key string) error {
	encrypted, err := encryption.EncryptAES([]byte(key), plainText)
	if err != nil {
		return err
	}

	message := &models.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    encrypted,
		Encrypted:  true,
	}
	return s.Repo.CreateMessage(message)
}

func (s *MessageService) GetMessages(userID uint, key string) ([]models.Message, error) {
	messages, err := s.Repo.GetMessagesForUser(userID)
	if err != nil {
		return nil, err
	}

	for i, msg := range messages {
		if msg.Encrypted {
			decrypted, err := encryption.DecryptAES([]byte(key), msg.Content)
			if err == nil {
				messages[i].Content = decrypted
			}
		}
	}
	return messages, nil
}

func (s *MessageService) DeleteMessage(messageID uint, userID uint) error {
	return s.Repo.DeleteMessage(messageID, userID)
}
