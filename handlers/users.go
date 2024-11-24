package handlers

import (
	"Animals_Shelter/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"html/template"
	"io"
	"log"
	"mime/multipart"
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	// Проверяем уникальность email
	var existingUser models.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		http.Error(w, "Email is already taken", http.StatusBadRequest)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking for existing email: %v", err)
		http.Error(w, "Error checking email", http.StatusInternalServerError)
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Error generating hashed password", http.StatusInternalServerError)
		return
	}

	// Преобразуем строку даты рождения в time.Time
	dateOfBirth, err := time.Parse("2006-01-02", dateOfBirthStr)
	if err != nil {
		http.Error(w, "Invalid date of birth format", http.StatusBadRequest)
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
			http.Error(w, "Unable to create directory", http.StatusInternalServerError)
			return
		}

		out, err := os.Create(profileImagePath)
		if err != nil {
			log.Printf("Error saving profile image: %v", err)
			http.Error(w, "Error saving profile image", http.StatusInternalServerError)
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
			http.Error(w, "Error saving profile image", http.StatusInternalServerError)
			return
		}
	} else {
		profileImagePath = "system_images/default_profile.jpg"
	}

	// Генерируем токен подтверждения email
	token, err := generateToken()
	if err != nil {
		log.Printf("Error generating token: %+v\n", err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	log.Printf("generated token: %s", token)

	// Начинаем транзакцию
	err = db.Transaction(func(tx *gorm.DB) error {
		// Вставляем данные в таблицу `users`
		user := models.User{
			Username: username,
			Password: string(hashedPassword),
			Email:    email,
			RoleID:   1, // Роль "User"
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// Вставляем данные в таблицу `user_details`
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

		// Вставляем данные в таблицу `user_images`
		userImage := models.UserImage{
			UserID:         user.ID,
			ProfileImage:   profileImagePath,
			ProfileBgImage: "system_images/default_bg.jpg",
		}
		if err := tx.Create(&userImage).Error; err != nil {
			return err
		}

		// Вставляем данные в таблицу `user_privacy`
		userPrivacy := models.UserPrivacy{
			UserID:    user.ID,
			ShowEmail: showEmail,
			ShowPhone: showPhone,
		}
		if err := tx.Create(&userPrivacy).Error; err != nil {
			return err
		}

		// Вставляем данные в таблицу `user_email_confirmations`
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
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	// Отправляем email для подтверждения
	if err := sendConfirmationEmail(email, token); err != nil {
		log.Printf("Error sending confirmation email: %v", err)
		http.Error(w, "Error sending confirmation email", http.StatusInternalServerError)
		return
	}

	// Перенаправляем на страницу успешной регистрации
	http.Redirect(w, r, "/registration_success", http.StatusSeeOther)
}

// Login handles user login using GORM and prepared statements
func Login(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Извлекаем данные пользователя
	var user models.User
	err := db.Where("username = ?", username).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		} else {
			errWrapped := fmt.Errorf("failed to query user data: %w", err)
			log.Printf("Login error: %+v\n", errWrapped)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Проверка на подтверждение email
	var emailConfirmation models.UserEmailConfirmation
	err = db.Where("user_id = ?", user.ID).First(&emailConfirmation).Error

	if err != nil {
		errWrapped := fmt.Errorf("failed to query email confirmation status: %w", err)
		log.Printf("Login error: %+v\n", errWrapped)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Проверка на подтверждение email
	if !emailConfirmation.EmailConfirmed {
		http.Error(w, "Please confirm your email address first", http.StatusForbidden)
		return
	}

	// Сравнение пароля с сохраненным хешем
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		errWrapped := fmt.Errorf("failed to compare password hash: %w", err)
		log.Printf("Login error: %+v\n", errWrapped)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Генерация сессионного токена
	sessionToken, err := generateToken()
	if err != nil {
		log.Printf("Error generating token: %+v\n", err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	// Сохраняем сессионный токен в базу данных
	session := models.Session{
		SessionID: sessionToken,
		UserID:    int(user.ID),
		ExpiresAt: &expiresAt,
	}
	if err := db.Create(&session).Error; err != nil {
		errWrapped := fmt.Errorf("failed to insert session into database: %w", err)
		log.Printf("Login error: %+v\n", errWrapped)
		http.Error(w, fmt.Sprintf("Error creating session: %v", err), http.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie с сессионным токеном
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   sessionToken,
		Expires: expiresAt,
	})

	// Перенаправляем на главную страницу после успешного логина
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	m := gomail.NewMessage()
	m.SetHeader("From", "mama.ne.7.gorui@gmail.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Please confirm your email address")
	m.SetBody("text/html", fmt.Sprintf("Click the link to confirm your email address: <a href='http://localhost:8080/confirm?token=%s'>Confirm Email</a>", token))

	d := gomail.NewDialer("smtp.gmail.com", 587, "mama.ne.7.gorui@gmail.com", "cuxw pvxk epvp yahf")
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
