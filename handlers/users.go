package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"html/template"
	"net/http"
	"regexp"
	"time"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	templates  = template.Must(template.ParseFiles("templates/register.html", "templates/login.html"))
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
}

// ShowRegisterForm renders the registration form
func ShowRegisterForm(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", nil)
}

// ShowLoginForm renders the login form
func ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", nil)
}

// Register handles user registration
func Register(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	email := r.FormValue("email")

	if !emailRegex.MatchString(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if len(password) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	if password != confirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating hashed password: %v", err), http.StatusInternalServerError)
		return
	}

	token := generateToken()

	// Insert user with role 'User' by default
	_, err = db.Exec("INSERT INTO users (username, password, email, email_confirmed, role, confirmation_token) VALUES ($1, $2, $3, $4, $5, $6)", username, string(hashedPassword), email, false, "User", token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting user into database: %v", err), http.StatusInternalServerError)
		return
	}

	if err := sendConfirmationEmail(email, token); err != nil {
		http.Error(w, fmt.Sprintf("Error sending confirmation email: %v", err), http.StatusInternalServerError)
		return
	}

	// Return a success message
	fmt.Fprintf(w, "Registration successful! Please check your email to confirm your account.")
}

// Login handles user login
func Login(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	err := db.QueryRow("SELECT id, password, email_confirmed FROM users WHERE username = $1", username).Scan(&user.ID, &user.Password, &user.EmailConfirmed)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if !user.EmailConfirmed {
		http.Error(w, "Please confirm your email address first", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Устанавливаем cookie сессии
	sessionToken := generateToken()
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   sessionToken,
		Expires: time.Now().Add(24 * time.Hour),
	})

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout handles user logout
func Logout(w http.ResponseWriter, r *http.Request) {
	// Удаляем cookie сессии
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour), // Устанавливаем прошедшее время для удаления cookie
	})

	// Перенаправляем на главную страницу
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

	fmt.Fprintf(w, "Email confirmed successfully!")
}
