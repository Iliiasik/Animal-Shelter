package auth

import (
	"Animals_Shelter/models"
	"gorm.io/gorm"
	"net/http"
)

// IsLoggedIn проверяет, залогинен ли пользователь
func IsLoggedIn(db *gorm.DB, r *http.Request) bool {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	var session models.Session
	err = db.Where("session_id = ?", sessionCookie.Value).First(&session).Error
	return err == nil
}

// IsAdmin проверяет, является ли текущий пользователь администратором
func IsAdmin(db *gorm.DB, r *http.Request) bool {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	var session models.Session
	err = db.Where("session_id = ?", sessionCookie.Value).First(&session).Error
	if err != nil {
		return false
	}

	var isAdmin bool
	stmt := db.Session(&gorm.Session{PrepareStmt: true}).Model(&models.User{}).
		Select("is_admin").Where("id = ?", session.UserID)
	err = stmt.Row().Scan(&isAdmin)
	return err == nil && isAdmin
}
