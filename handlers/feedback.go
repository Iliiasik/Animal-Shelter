package handlers

import (
	"Animals_Shelter/models"
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"

	"log"
	"net/http"
	"time"
)

type Session struct {
	ID     uint   `gorm:"primaryKey"`
	Token  string `gorm:"uniqueIndex"`
	UserID uint
}

func SaveFeedback(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			log.Println("Invalid method:", r.Method)
			return
		}

		// Чтение данных от клиента
		text := r.FormValue("text")
		if text == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Feedback text is required"})
			return
		}

		log.Printf("Received feedback: %s\n", text)

		// Проверка токена сессии
		sessionToken, err := getSessionToken(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Unauthorized"})
			log.Println("Error retrieving session cookie:", err)
			return
		}

		// Получаем ID пользователя по сессии
		userID, err := getUserIDBySession(db, sessionToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Unauthorized"})
			log.Println("Error getting user ID:", err)
			return
		}

		// Проверяем количество отзывов за последнюю неделю
		if isLimitExceeded(db, userID) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "You can only leave 2 feedbacks per week"})
			return
		}

		// Сохраняем отзыв в базу
		feedback := models.Feedback{
			Text:   text,
			UserID: userID,
		}

		err = db.Create(&feedback).Error
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Failed to save feedback"})
			log.Println("Error saving feedback to database:", err)
			return
		}

		log.Println("Feedback saved successfully")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "All set! Your feedback means the world to us."})
	}
}

// Функция для проверки, если лимит отзывов на текущую неделю превышен
func isLimitExceeded(db *gorm.DB, userID uint) bool {
	// Определяем время 7 дней назад от текущего времени
	oneWeekAgo := time.Now().AddDate(0, 0, -7)

	var feedbackCount int64
	// Считаем количество отзывов пользователя за последние 7 дней
	err := db.Model(&models.Feedback{}).Where("user_id = ? AND created_at >= ?", userID, oneWeekAgo).Count(&feedbackCount).Error
	if err != nil {
		log.Println("Error counting feedbacks:", err)
		return false // В случае ошибки не ограничиваем, чтобы избежать блокировки
	}
	log.Println("Feedback count", feedbackCount)
	// Возвращаем true, если отзывов больше или равно 2
	return feedbackCount >= 2
}

// getUserIDBySession принимает токен сессии и возвращает связанный с ним ID пользователя.
func getUserIDBySession(db *gorm.DB, sessionToken string) (uint, error) {
	if sessionToken == "" {
		return 0, errors.New("session token is empty")
	}

	var session Session
	// Ищем сессию в базе данных
	err := db.Where("session_id = ?", sessionToken).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("session not found")
		}
		return 0, err
	}

	return session.UserID, nil
}

// LoadFeedbackPage Метод для загрузки страницы с формой отзыва
func LoadFeedbackPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Загружаем HTML-шаблон страницы отзыва
	http.ServeFile(w, r, "templates/feedback.html")
}
