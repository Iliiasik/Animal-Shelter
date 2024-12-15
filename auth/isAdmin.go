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

// IsRole проверяет, имеет ли текущий пользователь определенную роль
func IsRole(db *gorm.DB, r *http.Request, roleID int) bool {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	// Найти сессию по идентификатору
	var session models.Session
	err = db.Where("session_id = ?", sessionCookie.Value).First(&session).Error
	if err != nil {
		return false
	}

	// Проверить роль пользователя
	var userRoleID int
	stmt := db.Session(&gorm.Session{PrepareStmt: true}).Model(&models.User{}).
		Select("role_id").Where("id = ?", session.UserID)
	err = stmt.Row().Scan(&userRoleID)

	return err == nil && userRoleID == roleID
}

// IsAdmin проверяет, является ли текущий пользователь администратором (role_id = 4)
func IsAdmin(db *gorm.DB, r *http.Request) bool {
	return IsRole(db, r, 4)
}
