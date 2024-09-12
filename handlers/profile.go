package handlers

import (
	"Animals_Shelter/models" // Путь к модели User
	"database/sql"
	"html/template"
	"net/http"
)

var profileTemplate = template.Must(template.ParseFiles("templates/profile.html"))

// ShowProfile показывает страницу профиля
func ShowProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор пользователя (например, из сессии)
	session, err := r.Cookie("session")
	if err != nil || session.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	err = db.QueryRow("SELECT id, username, email, first_name, last_name, bio, profile_image, phone_number, date_of_birth, city, country, website FROM users WHERE session_token = $1", session.Value).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Bio, &user.ProfileImage, &user.PhoneNumber, &user.DateOfBirth, &user.City, &user.Country, &user.Website)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Отображаем шаблон с данными пользователя
	err = profileTemplate.Execute(w, user)
	if err != nil {
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
	}
}
