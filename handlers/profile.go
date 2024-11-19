package handlers

import (
	"Animals_Shelter/models" // Путь к модели User и Session
	_ "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"mime/multipart"

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

// ShowProfile Handles rendering the profile of current user
func ShowProfile(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err != nil || sessionCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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

	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Session expired", http.StatusUnauthorized)
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

	err = profileTemplate.Execute(w, user)
	if err != nil {
		log.Println("Template rendering error:", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
	}
}

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

	if !user.DateOfBirth.IsZero() {
		user.FormattedDateOfBirth = user.DateOfBirth.Format("2006-01-02")
	}

	log.Printf("Profile background image URL: %s", user.ProfileBgImage)

	err = templates.ExecuteTemplate(w, "edit_profile.html", user)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// SaveProfile handles saving the updated user profile including the cropped image
func SaveProfile(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Println("Invalid method:", r.Method)
		return
	}

	firstName, lastName, bio, phone, dob, removeProfileImage, removeBackgroundImage := getFormData(r)
	log.Printf("Form data: firstName=%s, lastName=%s, bio=%s, phone=%s, dob=%s, removeProfileImage=%t, removeBackgroundImage=%t\n",
		firstName, lastName, bio, phone, dob, removeProfileImage, removeBackgroundImage)

	sessionToken, err := getSessionToken(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		log.Println("Error retrieving session cookie:", err)
		return
	}

	// Get existing images
	oldProfileImagePath, oldBackgroundImagePath, err := getUserImages(db, sessionToken)
	if err != nil {
		log.Println("Error fetching old images:", err)
		return
	}

	profileImagePath := handleImageUpdate(r, "croppedImage", "uploads/profile_images", removeProfileImage, oldProfileImagePath, defaultProfileImagePath)

	backgroundImagePath := handleImageUpdate(r, "backgroundImage", "uploads/profile_images/background", removeBackgroundImage, oldBackgroundImagePath, defaultBackgroundImagePath)

	// Updating profile
	err = updateUserProfile(db, firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath, sessionToken)
	if err != nil {
		http.Error(w, "Error saving profile", http.StatusInternalServerError)
		log.Println("Error executing query:", err)
		return
	}

	log.Println("Profile successfully updated")
	w.Header().Set("Content-Type", "application/json")
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
func getUserImages(db *gorm.DB, sessionToken string) (string, string, error) {
	var userID uint
	if err := db.Table("sessions").Select("user_id").Where("session_id = ?", sessionToken).Scan(&userID).Error; err != nil {
		return "", "", err
	}

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		return "", "", err
	}

	return user.ProfileImage, user.ProfileBgImage, nil
}

func handleImageUpdate(r *http.Request, formFieldName, dir string, removeFlag bool, oldImagePath, defaultImagePath string) string {
	if removeFlag {
		//Delete image if it is not system image
		if oldImagePath != "" && oldImagePath != defaultImagePath {
			if err := os.Remove(oldImagePath); err != nil {
				log.Println("Error deleting old image file:", err)
			} else {
				log.Printf("Deleted old image: %s", oldImagePath)
			}
		}
		return defaultImagePath
	}

	newImagePath, err := saveImage(r, formFieldName, dir)
	if err == nil && newImagePath != "" {
		//If new image uploaded, delete previous one
		if oldImagePath != "" && oldImagePath != defaultImagePath {
			if err := os.Remove(oldImagePath); err != nil {
				log.Println("Error deleting old image file:", err)
			}
		}
		return newImagePath
	}

	return oldImagePath
}
func updateUserProfile(db *gorm.DB, firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath, sessionToken string) error {
	// Find user ID from session token
	var userID uint
	if err := db.Table("sessions").Select("user_id").Where("session_id = ?", sessionToken).Scan(&userID).Error; err != nil {
		return err
	}

	// Update user profile using GORM
	if err := db.Model(&User{}).Where("id = ?", userID).Updates(User{
		FirstName:      firstName,
		LastName:       lastName,
		Bio:            bio,
		PhoneNumber:    phone,
		DateOfBirth:    dob,
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

	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		log.Println("Error retrieving session cookie:", err)
		return
	}

	sessionToken := cookie.Value
	log.Println("Session token:", sessionToken)

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

	if err := db.Model(&models.User{}).Where("id = ?", session.UserID).Updates(map[string]interface{}{
		"show_email": showEmail,
		"show_phone": showPhone,
	}).Error; err != nil {
		http.Error(w, "Error updating visibility settings", http.StatusInternalServerError)
		log.Println("Error executing query:", err)
		return
	}

	log.Println("Visibility settings successfully updated")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]bool{"success": true})
	if err != nil {
		log.Println("Error encoding JSON response:", err)
	}
}

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

	log.Println("Profile image path:", user.ProfileImage)
	log.Println("Background image path:", user.ProfileBgImage)

	err := userProfile.Execute(w, user)
	if err != nil {
		log.Println("Template rendering error:", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
	}
}
