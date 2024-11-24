package handlers

import (
	"Animals_Shelter/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func SaveFeedback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var feedback models.Feedback
		if err := c.ShouldBindJSON(&feedback); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Предполагаем, что текущий пользователь сохранен в сессии
		userID, exists := c.Get("userID") // получаем ID пользователя из контекста (сессии)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		feedback.UserID = userID.(uint) // устанавливаем ID пользователя, который оставил отзыв

		// Сохраняем отзыв в базе данных
		if err := db.Create(&feedback).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save feedback"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Feedback saved successfully"})
	}
}
