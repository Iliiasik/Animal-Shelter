package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	templates  = template.Must(template.ParseFiles("templates/register.html", "templates/login.html", "templates/edit_profile.html"))
)

// User represents a user in the database
type User struct {
	ID                int
	Username          string
	Password          string
	Email             string
	IsAdmin           bool
	EmailConfirmed    bool
	ConfirmationToken string
	FirstName         string
	LastName          string
	Bio               string
	ProfileImage      string
	PhoneNumber       string
	DateOfBirth       string
}

// ShowRegisterForm renders the registration form
func ShowRegisterForm(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", nil)
}

// ShowLoginForm renders the login form
func ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", nil)
}
func Register(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderError(w, r, "Method not allowed")
		return
	}

	// Получаем данные из формы
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	bio := r.FormValue("bio")
	phoneNumber := r.FormValue("phone_number")
	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	email := r.FormValue("email")
	dateOfBirth := r.FormValue("date_of_birth")

	// Проверяем корректность полей
	if !emailRegex.MatchString(email) {
		renderError(w, r, "Invalid email format")
		return
	}

	if len(password) < 8 {
		renderError(w, r, "Password must be at least 8 characters long")
		return
	}

	if password != confirmPassword {
		renderError(w, r, "Passwords do not match")
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		renderError(w, r, fmt.Sprintf("Error generating hashed password: %v", err))
		return
	}

	// Обрабатываем загрузку изображения
	file, header, err := r.FormFile("profile_image")
	if err != nil && err != http.ErrMissingFile {
		renderError(w, r, "Error uploading profile image")
		return
	}
	var imagePath string
	if file != nil {
		defer file.Close()
		imagePath = fmt.Sprintf("uploads/profile_images/%s", header.Filename)

		// Сохраняем изображение на диск
		out, err := os.Create(imagePath)
		if err != nil {
			renderError(w, r, "Error saving profile image")
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			renderError(w, r, "Error saving profile image")
			return
		}
	} else {
		imagePath = "" // Если изображение не загружено
	}

	// Генерируем токен подтверждения email
	token := generateToken()

	// Добавляем пользователя в базу данных
	_, err = db.Exec(`
		INSERT INTO users 
		(username, password, email, email_confirmed, role, confirmation_token, first_name, last_name, bio, phone_number, profile_image, date_of_birth) 
		VALUES 
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		username, string(hashedPassword), email, false, "User", token, firstName, lastName, bio, phoneNumber, imagePath, dateOfBirth)

	if err != nil {
		renderError(w, r, "Username or email is already taken")
		return
	}

	// Отправляем подтверждение на email
	if err := sendConfirmationEmail(email, token); err != nil {
		renderError(w, r, fmt.Sprintf("Error sending confirmation email: %v", err))
		return
	}

	// Перенаправляем пользователя на страницу успешной регистрации
	http.Redirect(w, r, "/registration_success", http.StatusSeeOther)
}

// Login handles user login
func Login(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderError(w, r, "Method not allowed")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	err := db.QueryRow("SELECT id, password, email_confirmed FROM users WHERE username = $1", username).Scan(&user.ID, &user.Password, &user.EmailConfirmed)
	if err != nil {
		if err == sql.ErrNoRows {
			renderError(w, r, "Invalid username or password")
		} else {
			renderError(w, r, "Internal server error")
		}
		return
	}

	if !user.EmailConfirmed {
		renderError(w, r, "Please confirm your email address first")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		renderError(w, r, "Invalid username or password")
		return
	}

	sessionToken := generateToken()

	// Save session token to the database
	_, err = db.Exec("INSERT INTO sessions (session_id, user_id, expires_at) VALUES ($1, $2, $3)", sessionToken, user.ID, time.Now().Add(24*time.Hour))
	if err != nil {
		renderError(w, r, fmt.Sprintf("Error creating session: %v", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   sessionToken,
		Expires: time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renderError(w http.ResponseWriter, r *http.Request, message string) {
	w.WriteHeader(http.StatusBadRequest)
	data := struct {
		ErrorMessage string
	}{
		ErrorMessage: message,
	}

	// Определяем какой шаблон использовать в зависимости от URL
	tmpl := "login.html"
	if r.URL.Path == "/register" {
		tmpl = "register.html"
	}

	templates.ExecuteTemplate(w, tmpl, data)
}

// Logout handles user logout
func Logout(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Get the session token from the cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessionToken := cookie.Value

	// Remove the session from the database
	_, err = db.Exec("DELETE FROM sessions WHERE session_id = $1", sessionToken)
	if err != nil {
		http.Error(w, "Error removing session", http.StatusInternalServerError)
		return
	}

	// Remove cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour), // Set an expired time to delete the cookie
	})

	// Redirect to the homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func sendConfirmationEmail(email, token string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "mama.ne.7.gorui@gmail.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Please confirm your email address")
	m.SetBody("text/html", fmt.Sprintf("Click the link to confirm your email address: <a href='http://localhost:8080/confirm?token=%s'>Confirm Email</a>", token))

	d := gomail.NewDialer("smtp.gmail.com", 587, "mama.ne.7.gorui@gmail.com", "cuxw pvxk epvp yahf")
	return d.DialAndSend(m)
}

// ConfirmEmail handles email confirmation
func ConfirmEmail(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE users SET email_confirmed = true WHERE confirmation_token = $1", token)
	if err != nil {
		http.Error(w, "Error confirming email", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error confirming email", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	tmpl, err := template.ParseFiles("templates/confirm.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
func EditProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получите ID пользователя из сессии или куки (здесь показан пример с куки)
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionToken := cookie.Value

	// Получите информацию о пользователе из базы данных (в примере предполагается, что у вас есть такой метод)
	var user User
	err = db.QueryRow("SELECT id, username, email, phone_number, bio, profile_image, date_of_birth FROM users WHERE id = (SELECT user_id FROM sessions WHERE session_id = $1)", sessionToken).Scan(&user.ID, &user.Username, &user.Email, &user.PhoneNumber, &user.Bio, &user.ProfileImage, &user.DateOfBirth)
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

// SaveProfile handles saving the updated user profile
func SaveProfile(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получите данные из формы
	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	bio := r.FormValue("bio")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	dob := r.FormValue("dob")

	// Получите ID пользователя из сессии (аналогично EditProfile)
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionToken := cookie.Value

	// Обновите информацию о пользователе в базе данных
	_, err = db.Exec("UPDATE users SET first_name = $1, last_name = $2, bio = $3, email = $4, phone_number = $5, date_of_birth = $6 WHERE id = (SELECT user_id FROM sessions WHERE session_id = $7)", firstName, lastName, bio, email, phone, dob, sessionToken)
	if err != nil {
		http.Error(w, "Error saving profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther) // Перенаправление на страницу профиля после сохранения
}
