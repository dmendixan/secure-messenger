package handlers

import (
	"gorm.io/gorm"
	"net/http"
	"secure-messenger/internal/models"

	"github.com/gin-gonic/gin"
)

func ProfileHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id") // ✅ достаем user_id из контекста

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		})
	}
}
