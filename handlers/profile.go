package handlers

import (
	"Animals_Shelter/models" // Путь к модели User и Session
	"database/sql"
	"html/template"
	"log"
	"net/http"
	_ "time"
)

var profileTemplate = template.Must(template.ParseFiles("templates/profile.html"))

// ShowProfile показывает страницу профиля
func ShowProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор сессии (например, из куки)
	sessionCookie, err := r.Cookie("session")
	if err != nil || sessionCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var session models.Session
	querySession := `
		SELECT id, session_id, user_id, created_at, expires_at 
		FROM sessions 
		WHERE session_id = $1
	`
	err = db.QueryRow(querySession, sessionCookie.Value).Scan(
		&session.ID, &session.SessionID, &session.UserID, &session.CreatedAt, &session.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Session not found", http.StatusUnauthorized)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Проверяем, не истекла ли сессия
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

	// Теперь получаем информацию о пользователе по user_id из сессии
	var user models.User
	queryUser := `
		SELECT id, username, email, first_name, last_name, bio, profile_image, phone_number, 
		       date_of_birth 
		FROM users 
		WHERE id = $1
	`
	err = db.QueryRow(queryUser, session.UserID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Bio,
		&user.ProfileImage, &user.PhoneNumber, &user.DateOfBirth,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Отображаем шаблон с данными пользователя
	err = profileTemplate.Execute(w, user)
	if err != nil {
		log.Println("Template rendering error:", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
	}
}
