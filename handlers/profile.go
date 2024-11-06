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

const defaultProfileImagePath = "system_images/default_profile.jpg"
const defaultBackgroundImagePath = "system_images/default_bg.jpg"

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
		       date_of_birth, profile_bg_image -- добавляем фоновое изображение
		FROM users 
		WHERE id = $1
	`
	err = db.QueryRow(queryUser, session.UserID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Bio,
		&user.ProfileImage, &user.PhoneNumber, &user.DateOfBirth,
		&user.ProfileBgImage, // добавляем поле для фона
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
        SELECT id, username, first_name, last_name, phone_number, bio, profile_image, date_of_birth, profile_bg_image
        FROM users 
        WHERE id = (SELECT user_id FROM sessions WHERE session_id = $1)`, sessionToken).
		Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Bio, &user.ProfileImage, &user.DateOfBirth, &user.ProfileBgImage)

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
	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	bio := r.FormValue("bio")
	phone := r.FormValue("phone")
	dob := r.FormValue("dob")

	// Флаги для удаления изображений
	removeProfileImage := r.FormValue("removeProfileImage") == "true"
	removeBackgroundImage := r.FormValue("removeBackgroundImage") == "true"

	log.Printf("Form data: firstName=%s, lastName=%s, bio=%s, phone=%s, dob=%s, removeProfileImage=%t, removeBackgroundImage=%t\n",
		firstName, lastName, bio, phone, dob, removeProfileImage, removeBackgroundImage)

	// Получение ID пользователя из сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		log.Println("Error retrieving session cookie:", err)
		return
	}
	sessionToken := cookie.Value

	// Получение старого изображения профиля и фона для удаления
	var oldProfileImagePath, oldBackgroundImagePath string
	err = db.QueryRow(`
        SELECT profile_image, profile_bg_image 
        FROM users 
        WHERE id = (SELECT user_id FROM sessions WHERE session_id = $1)`, sessionToken).
		Scan(&oldProfileImagePath, &oldBackgroundImagePath)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error fetching old images:", err)
	}

	// Загрузка новых изображений и удаление старых при необходимости
	var profileImagePath = oldProfileImagePath
	if removeProfileImage {
		profileImagePath = defaultProfileImagePath
		if oldProfileImagePath != "" && oldProfileImagePath != defaultProfileImagePath {
			err = os.Remove(oldProfileImagePath)
			if err != nil {
				log.Println("Error deleting old profile image file:", err)
			} else {
				log.Printf("Deleted old profile image: %s", oldProfileImagePath)
			}
		}
	} else {
		// Попытка загрузить новое изображение
		newProfileImagePath, err := saveImage(r, "croppedImage", "uploads/profile_images")
		if err == nil && newProfileImagePath != "" {
			// Удаляем старое изображение, если загрузилось новое
			if oldProfileImagePath != "" && oldProfileImagePath != defaultProfileImagePath {
				err = os.Remove(oldProfileImagePath)
				if err != nil {
					log.Println("Error deleting old profile image file:", err)
				}
			}
			profileImagePath = newProfileImagePath
			log.Printf("New profile image path: %s", profileImagePath)
		}
	}

	var backgroundImagePath = oldBackgroundImagePath
	if removeBackgroundImage {
		backgroundImagePath = defaultBackgroundImagePath
		if oldBackgroundImagePath != "" && oldBackgroundImagePath != defaultBackgroundImagePath {
			err = os.Remove(oldBackgroundImagePath)
			if err != nil {
				log.Println("Error deleting old background image file:", err)
			} else {
				log.Printf("Deleted old background image: %s", oldBackgroundImagePath)
			}
		}
	} else {
		// Попытка загрузить новое фоновое изображение
		newBackgroundImagePath, err := saveImage(r, "backgroundImage", "uploads/profile_images/background")
		if err == nil && newBackgroundImagePath != "" {
			// Удаляем старое фоновое изображение, если загрузилось новое
			if oldBackgroundImagePath != "" && oldBackgroundImagePath != defaultBackgroundImagePath {
				err = os.Remove(oldBackgroundImagePath)
				if err != nil {
					log.Println("Error deleting old background image file:", err)
				}
			}
			backgroundImagePath = newBackgroundImagePath
			log.Printf("New background image path: %s", backgroundImagePath)
		}
	}

	// Формирование запроса на обновление профиля пользователя
	query := `UPDATE users SET first_name = $1, last_name = $2, bio = $3, phone_number = $4, date_of_birth = $5, profile_image = $6, profile_bg_image = $7 WHERE id = (SELECT user_id FROM sessions WHERE session_id = $8)`
	params := []interface{}{firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath, sessionToken}

	_, err = db.Exec(query, params...)
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

// Функция для сохранения изображения в указанную директорию
func saveImage(r *http.Request, formFieldName, dir string) (string, error) {
	file, handler, err := r.FormFile(formFieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil // Файл не загружен
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
