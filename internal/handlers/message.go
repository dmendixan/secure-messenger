package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"secure-messenger/internal/services"
	"strconv"
)

type MessageHandler struct {
	Service *services.MessageService
}

func NewMessageHandler(s *services.MessageService) *MessageHandler {
	return &MessageHandler{Service: s}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req struct {
		ReceiverID uint   `json:"receiver_id"`
		Content    string `json:"content"`
		Key        string `json:"key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	userID := c.GetUint("user_id")
	err := h.Service.SendMessage(userID, req.ReceiverID, req.Content, req.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sent"})
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	key := c.Query("key")
	userID := c.GetUint("user_id")

	messages, err := h.Service.GetMessages(userID, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	userID := c.GetUint("user_id")

	err := h.Service.DeleteMessage(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
