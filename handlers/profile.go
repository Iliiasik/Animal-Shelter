package handlers

import (
	"Animals_Shelter/models" // Путь к модели User
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"
)

var profileTemplate = template.Must(template.ParseFiles("templates/profile.html"))

// ShowProfile показывает страницу профиля без использования модели Session
func ShowProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор сессии из куки
	sessionCookie, err := r.Cookie("session")
	if err != nil || sessionCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Проверяем сессию пользователя в базе данных
	var userID int
	var expiresAt time.Time
	querySession := `
		SELECT user_id, expires_at 
		FROM sessions 
		WHERE session_id = $1
	`
	err = db.QueryRow(querySession, sessionCookie.Value).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Session not found", http.StatusUnauthorized)
			return
		}
		log.Println("Database error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Проверяем, не истекла ли сессия
	if time.Now().After(expiresAt) {
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

	// Получаем информацию о пользователе по user_id из сессии
	var user models.User
	queryUser := `
		SELECT id, username, email, first_name, last_name, bio, profile_image, phone_number, 
		       date_of_birth, is_admin
		FROM users 
		WHERE id = $1
	`
	err = db.QueryRow(queryUser, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Bio,
		&user.ProfileImage, &user.PhoneNumber, &user.DateOfBirth, &user.IsAdmin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Println("Database error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// Отображаем шаблон с данными пользователя
	err = profileTemplate.Execute(w, user)
	if err != nil {
		log.Println("Template rendering error:", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
	}
}
