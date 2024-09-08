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
		renderError(w, r, "Method not allowed")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	email := r.FormValue("email")

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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		renderError(w, r, fmt.Sprintf("Error generating hashed password: %v", err))
		return
	}

	token := generateToken()

	_, err = db.Exec("INSERT INTO users (username, password, email, email_confirmed, role, confirmation_token) VALUES ($1, $2, $3, $4, $5, $6)", username, string(hashedPassword), email, false, "User", token)
	if err != nil {
		renderError(w, r, "Username or email is already taken")
		return
	}

	if err := sendConfirmationEmail(email, token); err != nil {
		renderError(w, r, fmt.Sprintf("Error sending confirmation email: %v", err))
		return
	}
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
