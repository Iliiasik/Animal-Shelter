package handlers

import (
	"Animals_Shelter/models"
	"Animals_Shelter/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

var profileTemplate = template.Must(template.ParseFiles("templates/profile.html", "templates/edit_profile.html"))

type UserProfile struct {
	User        models.User
	UserDetail  models.UserDetail
	UserImage   models.UserImage
	UserPrivacy models.UserPrivacy
}

const defaultProfileImagePath = "system_images/default_profile.jpg"
const defaultBackgroundImagePath = "system_images/default_bg.jpg"
const profileImageDir = "uploads/profile_images"
const backgroundImageDir = "uploads/profile_images/background"

// SaveProfile handles saving the updated user profile including the cropped image
func SaveProfile(db *gorm.DB, minioService *storage.MinioService, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Println("Invalid method:", r.Method)
		return
	}

	// Дополнительная проверка на сервере для предотвращения дублирующих запросов

	firstName, lastName, bio, phone, dob, removeProfileImage, removeBackgroundImage := getFormData(r)
	log.Printf("Form data: firstName=%s, lastName=%s, bio=%s, phone=%s, dob=%s, removeProfileImage=%t, removeBackgroundImage=%t\n",
		firstName, lastName, bio, phone, dob, removeProfileImage, removeBackgroundImage)

	sessionToken, err := getSessionToken(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		log.Println("Error retrieving session cookie:", err)
		return
	}

	var userID uint
	if err := db.Table("sessions").Select("user_id").Where("session_id = ?", sessionToken).Scan(&userID).Error; err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		log.Println("Error getting user ID from session:", err)
		return
	}

	var userDetail models.UserDetail
	var userImage models.UserImage
	if err := db.First(&userDetail, "user_id = ?", userID).Error; err != nil {
		log.Println("Error fetching user details:", err)
		return
	}

	if err := db.First(&userImage, "user_id = ?", userID).Error; err != nil {
		log.Println("Error fetching user images:", err)
		return
	}

	oldProfileImagePath := userImage.ProfileImage
	oldBackgroundImagePath := userImage.ProfileBgImage

	profileImagePath := handleImageUpdate(r, "croppedImage", "profile-images", minioService, removeProfileImage, oldProfileImagePath, defaultProfileImagePath)
	backgroundImagePath := handleImageUpdate(r, "backgroundImage", "background-images", minioService, removeBackgroundImage, oldBackgroundImagePath, defaultBackgroundImagePath)

	err = updateUserProfile(db, userID, firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath)
	if err != nil {
		http.Error(w, "Error saving profile", http.StatusInternalServerError)
		log.Println("Error executing query:", err)
		return
	}

	log.Println("Profile successfully updated")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]bool{"success": true})
	if err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func getFormData(r *http.Request) (string, string, string, string, string, bool, bool) {
	return r.FormValue("firstName"),
		r.FormValue("lastName"),
		r.FormValue("bio"),
		r.FormValue("phone"),
		r.FormValue("dob"),
		r.FormValue("removeProfileImage") == "true",
		r.FormValue("removeBackgroundImage") == "true"
}

func getSessionToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func handleImageUpdate(r *http.Request, formFieldName, bucketName string, minioService *storage.MinioService, removeFlag bool, oldImagePath, defaultImagePath string) string {
	if removeFlag {
		// Удаление изображения из MinIO, если оно не является системным изображением
		if oldImagePath != "" && oldImagePath != defaultImagePath {
			objectName := getObjectNameFromPath(oldImagePath)
			err := minioService.Client.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
			if err != nil {
				log.Println("Error deleting old image in MinIO:", err)
			} else {
				log.Printf("Deleted old image: %s from bucket %s", objectName, bucketName)
			}
		}
		return defaultImagePath
	}

	newImagePath, err := saveImageToMinIO(r, formFieldName, bucketName, minioService)
	if err == nil && newImagePath != "" {
		// Если загружено новое изображение, удалите предыдущее
		if oldImagePath != "" && oldImagePath != defaultImagePath {
			objectName := getObjectNameFromPath(oldImagePath)
			err := minioService.Client.RemoveObject(context.Background(), bucketName, objectName, minio.RemoveObjectOptions{})
			if err != nil {
				log.Println("Error deleting old image in MinIO:", err)
			}
		}
		return newImagePath
	}

	return oldImagePath
}
func saveImageToMinIO(r *http.Request, formFieldName, bucketName string, minioService *storage.MinioService) (string, error) {
	file, handler, err := r.FormFile(formFieldName)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", nil
		}
		log.Println("Error reading file:", err)
		return "", err
	}
	defer file.Close()

	// Уникальное имя файла
	objectName := storage.GenerateUniqueFileName(handler.Filename)

	// Загрузка файла в MinIO
	_, err = minioService.UploadFile(bucketName, objectName, file, handler.Size, handler.Header.Get("Content-Type"))
	if err != nil {
		log.Println("Error uploading file to MinIO:", err)
		return "", err
	}

	// Возвращаем путь к объекту
	return fmt.Sprintf("%s/%s", bucketName, objectName), nil
}

func getObjectNameFromPath(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func updateUserProfile(db *gorm.DB, userID uint, firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath string) error {
	// Update user detail fields
	dateOfBirth, _ := time.Parse("2006-01-02", dob)
	if err := db.Model(&models.UserDetail{}).Where("user_id = ?", userID).Updates(models.UserDetail{
		FirstName:   firstName,
		LastName:    lastName,
		Bio:         bio,
		PhoneNumber: phone,
		DateOfBirth: dateOfBirth,
	}).Error; err != nil {
		return err
	}

	// Update user image fields
	if err := db.Model(&models.UserImage{}).Where("user_id = ?", userID).Updates(models.UserImage{
		ProfileImage:   profileImagePath,
		ProfileBgImage: backgroundImagePath,
	}).Error; err != nil {
		return err
	}

	return nil
}

func saveImage(r *http.Request, formFieldName, dir string) (string, error) {
	file, handler, err := r.FormFile(formFieldName)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", nil
		}
		log.Println("Error uploading image:", err)
		return "", err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}(file)

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
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}(out)

	_, err = io.Copy(out, file)
	if err != nil {
		log.Println("Error copying file contents:", err)
		return "", err
	}

	return imagePath, nil
}

//ALl above are functions handling the update

// SaveVisibilitySettings Handles saving visibility settings from edit template (email,phone number)
func SaveVisibilitySettings(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to update visibility settings")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Println("Invalid method:", r.Method)
		return
	}

	showEmail := r.FormValue("showEmail") == "true"
	showPhone := r.FormValue("showPhone") == "true"

	log.Printf("Received form values: showEmail=%v, showPhone=%v", showEmail, showPhone)

	// Извлечение сессионного токена
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		log.Println("Error retrieving session cookie:", err)
		return
	}

	sessionToken := cookie.Value
	log.Println("Session token:", sessionToken)

	var session models.Session
	// Поиск сессии по токену
	if err := db.Where("session_id = ?", sessionToken).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Session not found", http.StatusUnauthorized)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Проверка существования настроек конфиденциальности для пользователя
	var privacy models.UserPrivacy
	if err := db.Where("user_id = ?", session.UserID).First(&privacy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если настройки конфиденциальности не найдены, создаем их
			privacy = models.UserPrivacy{
				UserID:    uint(session.UserID),
				ShowEmail: showEmail,
				ShowPhone: showPhone,
			}
			if err := db.Create(&privacy).Error; err != nil {
				http.Error(w, "Error saving visibility settings", http.StatusInternalServerError)
				log.Println("Error creating privacy settings:", err)
				return
			}
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		// Обновляем существующие настройки конфиденциальности
		if err := db.Model(&privacy).Updates(map[string]interface{}{
			"show_email": showEmail,
			"show_phone": showPhone,
		}).Error; err != nil {
			http.Error(w, "Error updating visibility settings", http.StatusInternalServerError)
			log.Println("Error executing query:", err)
			return
		}
	}

	log.Println("Visibility settings successfully updated")

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]bool{"success": true})
	if err != nil {
		log.Println("Error encoding JSON response:", err)
	}
}

// ShowProfile handles rendering the profile of the current user
func ShowProfile(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	// Проверяем наличие cookie сессии
	sessionCookie, err := r.Cookie("session")
	if err != nil || sessionCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Ищем сессию в базе данных
	var session models.Session
	if err := db.Where("session_id = ?", sessionCookie.Value).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Session not found", http.StatusUnauthorized)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Проверка на истечение срока действия сессии
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

	// Ищем пользователя по UserID из сессии
	var user models.User
	if err := db.Preload("Role").First(&user, session.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Загружаем дополнительные данные о пользователе (детали, изображения, конфиденциальность)
	var userDetail models.UserDetail
	if err := db.First(&userDetail, "user_id = ?", user.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("User details not found for user ID:", user.ID)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	var userImage models.UserImage
	if err := db.First(&userImage, "user_id = ?", user.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("User images not found for user ID:", user.ID)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	var userPrivacy models.UserPrivacy
	if err := db.First(&userPrivacy, "user_id = ?", user.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("User privacy settings not found for user ID:", user.ID)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Формируем структурированные данные для шаблона
	profileData := struct {
		User        models.User
		UserDetail  models.UserDetail
		UserImage   models.UserImage
		UserPrivacy models.UserPrivacy
	}{
		User:        user,
		UserDetail:  userDetail,
		UserImage:   userImage,
		UserPrivacy: userPrivacy,
	}

	// Отправляем данные в шаблон для рендеринга
	err = profileTemplate.Execute(w, profileData)
	if err != nil {
		log.Println("Template rendering error:", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
	}
}

// RenderEditTemplate Rendering the Edit template of current user with his existing info
// RenderEditTemplate Rendering the Edit template of current user with his existing info
func RenderEditTemplate(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionToken := cookie.Value

	var session models.Session
	if err := db.Where("session_id = ?", sessionToken).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Session not found", http.StatusUnauthorized)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	var user models.User
	if err := db.First(&user, session.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Загрузите дополнительные данные о пользователе
	var userDetail models.UserDetail
	if err := db.Where("user_id = ?", user.ID).First(&userDetail).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userDetail = models.UserDetail{}
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	var userImage models.UserImage
	if err := db.Where("user_id = ?", user.ID).First(&userImage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userImage = models.UserImage{}
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	var userPrivacy models.UserPrivacy
	if err := db.Where("user_id = ?", user.ID).First(&userPrivacy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userPrivacy = models.UserPrivacy{}
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Форматирование даты
	if !userDetail.DateOfBirth.IsZero() {
		userDetail.FormattedDateOfBirth = userDetail.DateOfBirth.Format("2006-01-02")
	}

	// Создание структуры для передачи в шаблон
	profile := UserProfile{
		User:        user,
		UserDetail:  userDetail,
		UserImage:   userImage,
		UserPrivacy: userPrivacy,
	}

	// Отправка данных в шаблон
	err = templates.ExecuteTemplate(w, "edit_profile.html", profile)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

var userProfile = template.Must(template.ParseFiles("templates/userProfile.html"))

// ViewProfile Handles rendering profiles of other users
func ViewProfile(db *gorm.DB, w http.ResponseWriter, username string) {
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Println("Database error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	var userDetails models.UserDetail
	if err := db.First(&userDetails, "user_id = ?", user.ID).Error; err != nil {
		log.Println("Error fetching user details:", err)
		return
	}

	var userImage models.UserImage
	if err := db.First(&userImage, "user_id = ?", user.ID).Error; err != nil {
		log.Println("Error fetching user images:", err)
		return
	}

	var userPrivacy models.UserPrivacy
	if err := db.First(&userPrivacy, "user_id = ?", user.ID).Error; err != nil {
		log.Println("Error fetching user privacy settings:", err)
		return
	}

	profile := struct {
		User        models.User
		UserDetail  models.UserDetail
		UserImage   models.UserImage
		UserPrivacy models.UserPrivacy
	}{
		User:        user,
		UserDetail:  userDetails,
		UserImage:   userImage,
		UserPrivacy: userPrivacy,
	}

	log.Println("Profile image path:", userImage.ProfileImage)
	log.Println("Background image path:", userImage.ProfileBgImage)

	err := userProfile.Execute(w, profile)
	if err != nil {
		log.Println("Template rendering error:", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
	}
}
