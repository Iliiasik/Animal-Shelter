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

// Handles rendering the profile of current user
func ShowProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {

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

	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Session expired", http.StatusUnauthorized)
		return
	}

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

	err = profileTemplate.Execute(w, user)
	if err != nil {
		log.Println("Template rendering error:", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
	}
}

// Rendering the Edit template of current user with his existing info
func RenderEditTemplate(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получите ID пользователя из сессии или куки
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionToken := cookie.Value

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

	if user.DateOfBirth != "" {
		parsedDate, err := time.Parse("2006-01-02T15:04:05Z", user.DateOfBirth)
		if err == nil {
			user.DateOfBirth = parsedDate.Format("2006-01-02") // Преобразуем в формат yyyy-MM-dd
		} else {
			log.Printf("Error parsing date of birth: %v", err)
		}
	}

	log.Printf("Profile background image URL: %s", user.ProfileBgImage)

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

	//Updating profile
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
func updateUserProfile(db *sql.DB, firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath, sessionToken string) error {
	query := `UPDATE users SET first_name = $1, last_name = $2, bio = $3, phone_number = $4, date_of_birth = $5, profile_image = $6, profile_bg_image = $7 WHERE id = (SELECT user_id FROM sessions WHERE session_id = $8)`
	params := []interface{}{firstName, lastName, bio, phone, dob, profileImagePath, backgroundImagePath, sessionToken}

	_, err := db.Exec(query, params...)
	return err
}
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

//ALl above are functions handling the update

// Handles saving visibility settings from edit template (email,phone number)
func SaveVisibilitySettings(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	query := `UPDATE users SET show_email = $1, show_phone = $2 WHERE id = (SELECT user_id FROM sessions WHERE session_id = $3)`
	_, err = db.Exec(query, showEmail, showPhone, sessionToken)
	if err != nil {
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

// Handles rendering profiles of other users
func ViewProfile(db *sql.DB, w http.ResponseWriter, r *http.Request, username string) {

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
