package handlers

import (
	"Animals_Shelter/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"

	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/jinzhu/gorm"
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
	EmailConfirmed    bool
	ConfirmationToken string
}

type UserDetails struct {
	UserID      int
	FirstName   string
	LastName    string
	Bio         string
	PhoneNumber string
	DateOfBirth string
}

type UserImages struct {
	UserID         int
	ProfileImage   string
	ProfileBgImage string
}

type UserPrivacy struct {
	UserID    int
	ShowEmail bool
	ShowPhone bool
}

// ShowRegisterForm renders the registration form
func ShowRegisterForm(w http.ResponseWriter) {
	err := templates.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		errWrapped := errors.Wrap(err, "error executing register template")
		log.Printf("ShowRegisterForm error: %+v\n", errWrapped)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ShowLoginForm renders the login form
func ShowLoginForm(w http.ResponseWriter) {
	err := templates.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		errWrapped := errors.Wrap(err, "error executing login template")
		log.Printf("ShowLoginForm error: %+v\n", errWrapped)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Register handles user registration using GORM and prepared statements
func Register(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
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
	dateOfBirthStr := r.FormValue("date_of_birth")
	showEmail := r.FormValue("show_email") == "true"
	showPhone := r.FormValue("show_phone") == "true"

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

	// Проверяем уникальность email
	var existingUser models.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		renderError(w, r, "Email is already taken")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking for existing email: %v", err)
		renderError(w, r, "Internal server error")
		return
	}
	var existingUsername models.User
	if err := db.Where("username = ?", username).First(&existingUsername).Error; err == nil {
		renderError(w, r, "Username is already taken")
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		renderError(w, r, "Internal server error")
		return
	}

	// Преобразуем строку даты рождения в time.Time
	dateOfBirth, err := time.Parse("2006-01-02", dateOfBirthStr)
	if err != nil {
		renderError(w, r, "Invalid date of birth format")
		return
	}

	// Обрабатываем загрузку изображения
	file, header, err := r.FormFile("profile_image")
	var profileImagePath string
	if err == nil && file != nil {
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				log.Println("Error closing file:", err)
			}
		}(file)

		uniqueFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
		profileImagePath = fmt.Sprintf("uploads/profile_images/%s", uniqueFileName)

		err := os.MkdirAll("uploads/profile_images", os.ModePerm)
		if err != nil {
			log.Printf("Error creating directory: %v", err)
			renderError(w, r, "Internal server error")
			return
		}

		out, err := os.Create(profileImagePath)
		if err != nil {
			log.Printf("Error saving profile image: %v", err)
			renderError(w, r, "Error saving profile image")
			return
		}
		defer func(out *os.File) {
			err := out.Close()
			if err != nil {
				log.Println("Error closing file:", err)
			}
		}(out)

		_, err = io.Copy(out, file)
		if err != nil {
			log.Printf("Error saving profile image: %v", err)
			renderError(w, r, "Error saving profile image")
			return
		}
	} else {
		profileImagePath = defaultProfileImagePath
	}

	// Генерируем токен подтверждения email
	token, err := generateToken()
	if err != nil {
		log.Printf("Error generating token: %+v\n", err)
		renderError(w, r, "Error generating token")
		return
	}

	// Начинаем транзакцию
	err = db.Transaction(func(tx *gorm.DB) error {
		user := models.User{
			Username: username,
			Password: string(hashedPassword),
			Email:    email,
			RoleID:   1,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		userDetail := models.UserDetail{
			UserID:      user.ID,
			FirstName:   firstName,
			LastName:    lastName,
			Bio:         bio,
			PhoneNumber: phoneNumber,
			DateOfBirth: dateOfBirth,
		}
		if err := tx.Create(&userDetail).Error; err != nil {
			return err
		}

		userImage := models.UserImage{
			UserID:         user.ID,
			ProfileImage:   profileImagePath,
			ProfileBgImage: "system_images/default_bg.jpg",
		}
		if err := tx.Create(&userImage).Error; err != nil {
			return err
		}

		userPrivacy := models.UserPrivacy{
			UserID:    user.ID,
			ShowEmail: showEmail,
			ShowPhone: showPhone,
		}
		if err := tx.Create(&userPrivacy).Error; err != nil {
			return err
		}

		emailConfirmation := models.UserEmailConfirmation{
			UserID:            user.ID,
			ConfirmationToken: token,
			EmailConfirmed:    false,
		}
		if err := tx.Create(&emailConfirmation).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Printf("Error registering user: %v", err)
		renderError(w, r, "Internal server error")
		return
	}

	// Отправляем email для подтверждения
	if err := sendConfirmationEmail(email, token); err != nil {
		log.Printf("Error sending confirmation email: %v", err)
		renderError(w, r, "Internal server error")
		return
	}

	http.Redirect(w, r, "/registration_success", http.StatusSeeOther)
}

// Login handles user login using GORM and prepared statements
func Login(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderError(w, r, "Method not allowed")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user models.User
	err := db.Where("username = ?", username).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			renderError(w, r, "Invalid username or password")
		} else {
			log.Printf("Login error: %v", err)
			renderError(w, r, "Internal server error")
		}
		return
	}

	var emailConfirmation models.UserEmailConfirmation
	err = db.Where("user_id = ?", user.ID).First(&emailConfirmation).Error

	if err != nil {
		log.Printf("Login error: %v", err)
		renderError(w, r, "Internal server error")
		return
	}

	if !emailConfirmation.EmailConfirmed {
		renderError(w, r, "Please confirm your email address first")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		renderError(w, r, "Invalid username or password")
		return
	}

	sessionToken, err := generateToken()
	if err != nil {
		log.Printf("Error generating token: %+v\n", err)
		renderError(w, r, "Internal server error")
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	session := models.Session{
		SessionID: sessionToken,
		UserID:    uint(user.ID),
		ExpiresAt: &expiresAt,
	}
	if err := db.Create(&session).Error; err != nil {
		log.Printf("Login error: %v", err)
		renderError(w, r, "Internal server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   sessionToken,
		Expires: expiresAt,
	})

	http.Redirect(w, r, "/?login=success", http.StatusSeeOther)
}

// Logout handles user logout using GORM
func Logout(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем сессионный токен из cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessionToken := cookie.Value

	// Удаляем сессию из базы данных
	if err := db.Where("session_id = ?", sessionToken).Delete(&models.Session{}).Error; err != nil {
		errWrapped := errors.Wrap(err, "failed to delete session from database")
		log.Printf("Logout error: %+v\n", errWrapped)
		http.Error(w, "Error removing session", http.StatusInternalServerError)
		return
	}

	// Удаляем cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour), // Устанавливаем истекшее время для удаления cookie
	})

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/?logout=success", http.StatusSeeOther)
}

// generateToken generates a secure token and handles any potential errors
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		errWrapped := errors.Wrap(err, "error generating token")
		log.Printf("generateToken error: %+v\n", errWrapped)
		return "", errWrapped
	}
	return hex.EncodeToString(b), nil
}

func sendConfirmationEmail(email, token string) error {
	err := godotenv.Load("configuration.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	if fromEmail == "" || password == "" || smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("missing required SMTP environment variables")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Please confirm your email address")
	m.SetBody("text/html", fmt.Sprintf("Click the link to confirm your email address: <a href='http://34.16.104.66:8080/confirm?token=%s'>Confirm Email</a>", token))

	// Создаем и настраиваем SMTP-диалер
	d := gomail.NewDialer(smtpHost, 587, fromEmail, password)

	// Отправляем письмо
	return d.DialAndSend(m)
}

// ConfirmEmail handles email confirmation using GORM and prepared statements
func ConfirmEmail(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		err := errors.New("invalid token provided")
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// Выполняем обновление подтверждения email
	result := db.Model(&models.UserEmailConfirmation{}).Where("confirmation_token = ?", token).Update("email_confirmed", true)
	if result.Error != nil {
		err := errors.Wrap(result.Error, "error executing update query for email confirmation")
		log.Printf("Error: %v", err)
		http.Error(w, "Error confirming email", http.StatusInternalServerError)
		return
	}

	// Проверка количества затронутых строк
	if result.RowsAffected == 0 {
		err := errors.New("no rows affected, invalid token")
		log.Printf("Error: %v", err)
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// Парсим шаблон
	tmpl, err := template.ParseFiles("templates/confirm.html")
	if err != nil {
		err = errors.Wrap(err, "error parsing confirmation template")
		log.Printf("Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Отправляем шаблон пользователю
	err = tmpl.Execute(w, nil)
	if err != nil {
		err = errors.Wrap(err, "error executing confirmation template")
		log.Printf("Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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

	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		errWrapped := errors.Wrap(err, "error executing register template")
		log.Printf("ShowRegisterForm error: %+v\n", errWrapped)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}
