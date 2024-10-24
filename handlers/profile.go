package handlers

import (
	"Animals_Shelter/models" // Путь к модели User и Session
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var profileTemplate = template.Must(template.ParseFiles("templates/profile.html", "templates/edit_profile.html"))

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
func EditProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получите ID пользователя из сессии или куки
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionToken := cookie.Value

	// Получите информацию о пользователе из базы данных
	var user User
	err = db.QueryRow(`
        SELECT id, username, first_name, last_name, phone_number, bio, profile_image, date_of_birth 
        FROM users 
        WHERE id = (SELECT user_id FROM sessions WHERE session_id = $1)`, sessionToken).
		Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Bio, &user.ProfileImage, &user.DateOfBirth)

	if err != nil {
		http.Error(w, "Error fetching user data", http.StatusInternalServerError)
		return
	}

	// Отправьте данные пользователя в шаблон
	err = templates.ExecuteTemplate(w, "edit_profile.html", user)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// SaveProfile handles saving the updated user profile including the cropped image
func SaveProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Println("Invalid method:", r.Method)
		return
	}

	// Получение данных из формы
	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	bio := r.FormValue("bio")
	phone := r.FormValue("phone")
	dob := r.FormValue("dob")

	// Логирование полученных данных формы
	log.Printf("Form data: firstName=%s, lastName=%s, bio=%s, phone=%s, dob=%s\n", firstName, lastName, bio, phone, dob)

	file, handler, err := r.FormFile("croppedImage")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Unable to upload image", http.StatusInternalServerError)
		log.Println("Error uploading image:", err)
		return
	}

	var imagePath string
	if file != nil {
		defer file.Close()

		// Генерация уникального имени файла
		uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
		imagePath = fmt.Sprintf("uploads/profile_images/%s", uniqueFileName)

		// Создаем директорию, если она не существует
		err := os.MkdirAll("uploads/profile_images", os.ModePerm)
		if err != nil {
			http.Error(w, "Unable to create directory", http.StatusInternalServerError)
			log.Println("Error creating directory:", err)
			return
		}

		// Сохраняем изображение на диск
		out, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Unable to save the image", http.StatusInternalServerError)
			log.Println("Error saving image:", err)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Error saving image", http.StatusInternalServerError)
			log.Println("Error copying file contents:", err)
			return
		}
	} else {
		log.Println("No image uploaded")
		imagePath = "" // Если изображение не загружено
	}

	// Получение ID пользователя из сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		log.Println("Error retrieving session cookie:", err)
		return
	}
	sessionToken := cookie.Value
	log.Println("Session token:", sessionToken)

	// Получение текущего пути к изображению пользователя из базы данных
	var currentImagePath string
	err = db.QueryRow(`SELECT profile_image FROM users WHERE id = (SELECT user_id FROM sessions WHERE session_id = $1)`, sessionToken).Scan(&currentImagePath)
	if err != nil {
		http.Error(w, "Error retrieving current profile image", http.StatusInternalServerError)
		log.Println("Error retrieving current image path:", err)
		return
	}

	// Удаление старого изображения, если оно существует
	if currentImagePath != "" {
		err := os.Remove(currentImagePath)
		if err != nil {
			log.Println("Error removing old profile image:", err)
		}
	}

	// Обновление профиля пользователя
	query := `UPDATE users SET first_name = $1, last_name = $2, bio = $3, phone_number = $4, date_of_birth = $5`
	if imagePath != "" {
		query += ", profile_image = $6 WHERE id = (SELECT user_id FROM sessions WHERE session_id = $7)"
		_, err = db.Exec(query, firstName, lastName, bio, phone, dob, imagePath, sessionToken)
		log.Println("Executing query with image")
	} else {
		query += " WHERE id = (SELECT user_id FROM sessions WHERE session_id = $6)"
		_, err = db.Exec(query, firstName, lastName, bio, phone, dob, sessionToken)
		log.Println("Executing query without image")
	}

	if err != nil {
		http.Error(w, "Error saving profile", http.StatusInternalServerError)
		log.Println("Error executing query:", err)
		return
	}

	log.Println("Profile successfully updated")
	// Возвращаем JSON ответ с информацией об успехе
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)                                // Установите статус 200 OK
	json.NewEncoder(w).Encode(map[string]bool{"success": true}) // Вернуть объект JSON с успешным статусом

	//http.Redirect(w, r, "/profile", http.StatusSeeOther)

}
