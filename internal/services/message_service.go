package services

import (
	"secure-messenger/internal/models"
	"secure-messenger/internal/repository"
	"secure-messenger/pkg/encryption"
)

type MessageService struct {
	Repo         *repository.MessageRepository
	AESSecretKey []byte
}

func NewMessageService(r *repository.MessageRepository, key []byte) *MessageService {
	return &MessageService{
		Repo:         r,
		AESSecretKey: key,
	}
}

func (s *MessageService) SendMessage(senderID, receiverID uint, plainText string) error {
	encrypted, err := encryption.EncryptAES(s.AESSecretKey, plainText)
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

func (s *MessageService) GetMessages(userID uint) ([]models.Message, error) {
	messages, err := s.Repo.GetMessagesForUser(userID)
	if err != nil {
		return nil, err
	}

	for i, msg := range messages {
		if msg.Encrypted {
			decrypted, err := encryption.DecryptAES(s.AESSecretKey, msg.Content)
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
