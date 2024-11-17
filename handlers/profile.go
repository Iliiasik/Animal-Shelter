package handlers

import (
	"Animals_Shelter/models" // Путь к модели User и Session
	"database/sql"
	"encoding/json"
	"fmt"
	/*"github.com/gorilla/mux"*/
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var profileTemplate = template.Must(template.ParseFiles("templates/profile.html", "templates/edit_profile.html"))
var userProfile = template.Must(template.ParseFiles("templates/userProfile.html"))

type Profile struct {
	ID           string
	FirstName    string
	LastName     string
	Email        string
	ProfileImage string
}

const defaultProfileImagePath = "system_images/default_profile.jpg"
const defaultBackgroundImagePath = "system_images/default_bg.jpg"

func ShowProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор сессии (например, из куки)
	sessionCookie, err := r.Cookie("session")
	if err != nil || sessionCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var session models.Session
	querySession := `SELECT id, session_id, user_id, created_at, expires_at FROM sessions WHERE session_id = $1`
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
		SELECT id, username, email, first_name, last_name, bio, profile_image, phone_number, date_of_birth, profile_bg_image, show_email, show_phone
		FROM users WHERE id = $1
	`
	err = db.QueryRow(queryUser, session.UserID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Bio,
		&user.ProfileImage, &user.PhoneNumber, &user.DateOfBirth, &user.ProfileBgImage,
		&user.ShowEmail, &user.ShowPhone,
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

func RenderEditTemplate(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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
        SELECT id, username, first_name, last_name, phone_number, bio, profile_image, date_of_birth, profile_bg_image,show_email, show_phone
        FROM users 
        WHERE id = (SELECT user_id FROM sessions WHERE session_id = $1)`, sessionToken).
		Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Bio, &user.ProfileImage, &user.DateOfBirth, &user.ProfileBgImage, &user.ShowEmail, &user.ShowPhone)

	if err != nil {
		http.Error(w, "Error fetching user data", http.StatusInternalServerError)
		return
	}

	// Преобразование строки даты в формат yyyy-MM-dd
	if user.DateOfBirth != "" {
		parsedDate, err := time.Parse("2006-01-02T15:04:05Z", user.DateOfBirth)
		if err == nil {
			user.DateOfBirth = parsedDate.Format("2006-01-02") // Преобразуем в формат yyyy-MM-dd
		} else {
			log.Printf("Error parsing date of birth: %v", err)
		}
	}

	// Логируем URL фонового изображения
	log.Printf("Profile background image URL: %s", user.ProfileBgImage)

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
	firstName, lastName, bio, phone, dob, removeProfileImage, removeBackgroundImage := getFormData(r)
	log.Printf("Form data: firstName=%s, lastName=%s, bio=%s, phone=%s, dob=%s, removeProfileImage=%t, removeBackgroundImage=%t\n",
		firstName, lastName, bio, phone, dob, removeProfileImage, removeBackgroundImage)

	// Получение ID пользователя из сессии
	sessionToken, err := getSessionToken(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		log.Println("Error retrieving session cookie:", err)
		return
	}

	// Получение старых изображений
	oldProfileImagePath, oldBackgroundImagePath, err := getUserImages(db, sessionToken)
	if err != nil {
		log.Println("Error fetching old images:", err)
		return
	}

	// Обработка изображения профиля
	profileImagePath := handleImageUpdate(r, "croppedImage", "uploads/profile_images", removeProfileImage, oldProfileImagePath, defaultProfileImagePath)

	// Обработка фонового изображения
	backgroundImagePath := handleImageUpdate(r, "backgroundImage", "uploads/profile_images/background", removeBackgroundImage, oldBackgroundImagePath, defaultBackgroundImagePath)

	// Обновление профиля
	err = updateUserProfile(db, firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath, sessionToken)
	if err != nil {
		http.Error(w, "Error saving profile", http.StatusInternalServerError)
		log.Println("Error executing query:", err)
		return
	}

	log.Println("Profile successfully updated")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Получение данных из формы
func getFormData(r *http.Request) (string, string, string, string, string, bool, bool) {
	return r.FormValue("firstName"),
		r.FormValue("lastName"),
		r.FormValue("bio"),
		r.FormValue("phone"),
		r.FormValue("dob"),
		r.FormValue("removeProfileImage") == "true",
		r.FormValue("removeBackgroundImage") == "true"
}

// Получение токена сессии
func getSessionToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// Получение старых изображений пользователя
func getUserImages(db *sql.DB, sessionToken string) (string, string, error) {
	var profileImagePath, backgroundImagePath string
	err := db.QueryRow(`
        SELECT profile_image, profile_bg_image 
        FROM users 
        WHERE id = (SELECT user_id FROM sessions WHERE session_id = $1)`, sessionToken).
		Scan(&profileImagePath, &backgroundImagePath)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	return profileImagePath, backgroundImagePath, err
}

// Обработка обновления изображения
func handleImageUpdate(r *http.Request, formFieldName, dir string, removeFlag bool, oldImagePath, defaultImagePath string) string {
	if removeFlag {
		// Удаление старого изображения, если оно не является изображением по умолчанию
		if oldImagePath != "" && oldImagePath != defaultImagePath {
			if err := os.Remove(oldImagePath); err != nil {
				log.Println("Error deleting old image file:", err)
			} else {
				log.Printf("Deleted old image: %s", oldImagePath)
			}
		}
		return defaultImagePath
	}

	// Загрузка нового изображения
	newImagePath, err := saveImage(r, formFieldName, dir)
	if err == nil && newImagePath != "" {
		// Удаление старого изображения, если новое загружено успешно
		if oldImagePath != "" && oldImagePath != defaultImagePath {
			if err := os.Remove(oldImagePath); err != nil {
				log.Println("Error deleting old image file:", err)
			}
		}
		return newImagePath
	}

	return oldImagePath
}

// Обновление профиля пользователя
func updateUserProfile(db *sql.DB, firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath, sessionToken string) error {
	query := `UPDATE users SET first_name = $1, last_name = $2, bio = $3, phone_number = $4, date_of_birth = $5, profile_image = $6, profile_bg_image = $7 WHERE id = (SELECT user_id FROM sessions WHERE session_id = $8)`
	params := []interface{}{firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath, sessionToken}

	_, err := db.Exec(query, params...)
	return err
}

// Функция для сохранения изображения
func saveImage(r *http.Request, formFieldName, dir string) (string, error) {
	file, handler, err := r.FormFile(formFieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		log.Println("Error uploading image:", err)
		return "", err
	}
	defer file.Close()

	uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
	imagePath := fmt.Sprintf("%s/%s", dir, uniqueFileName)

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Println("Error creating directory:", err)
		return "", err
	}

	out, err := os.Create(imagePath)
	if err != nil {
		log.Println("Error saving image:", err)
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		log.Println("Error copying file contents:", err)
		return "", err
	}

	return imagePath, nil
}

func SaveVisibilitySettings(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to update visibility settings")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Println("Invalid method:", r.Method)
		return
	}

	// Получение данных из формы
	showEmail := r.FormValue("showEmail") == "true"
	showPhone := r.FormValue("showPhone") == "true"

	// Логирование данных, полученных из формы
	log.Printf("Received form values: showEmail=%v, showPhone=%v", showEmail, showPhone)

	// Получение ID пользователя из сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		log.Println("Error retrieving session cookie:", err)
		return
	}

	sessionToken := cookie.Value
	log.Println("Session token:", sessionToken)

	// Обновление настроек видимости в базе данных
	query := `UPDATE users SET show_email = $1, show_phone = $2 WHERE id = (SELECT user_id FROM sessions WHERE session_id = $3)`
	_, err = db.Exec(query, showEmail, showPhone, sessionToken)
	if err != nil {
		http.Error(w, "Error updating visibility settings", http.StatusInternalServerError)
		log.Println("Error executing query:", err)
		return
	}

	log.Println("Visibility settings successfully updated")

	// Ответ клиенту об успешном обновлении
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]bool{"success": true})
	if err != nil {
		log.Println("Error encoding JSON response:", err)
	}
}
func ViewProfile(db *sql.DB, w http.ResponseWriter, r *http.Request, username string) {
	// Получаем информацию о пользователе по username
	var user models.User
	queryUser := `
        SELECT id, username, email, first_name, last_name, bio, profile_image, phone_number, date_of_birth, profile_bg_image, show_email, show_phone
        FROM users WHERE username = $1
    `
	err := db.QueryRow(queryUser, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Bio,
		&user.ProfileImage, &user.PhoneNumber, &user.DateOfBirth, &user.ProfileBgImage,
		&user.ShowEmail, &user.ShowPhone,
	)
	log.Println("Profile image path:", user.ProfileImage)
	log.Println("Background image path:", user.ProfileBgImage)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	err = userProfile.Execute(w, user)

	if err != nil {
		log.Println("Template rendering error:", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
	}
}
