package auth

import (
	"Animals_Shelter/models"
	"net/http"

	"github.com/jinzhu/gorm"
)

func IsLoggedIn(db *gorm.DB, r *http.Request) bool {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	var session models.Session
	err = db.Where("session_id = ?", sessionCookie.Value).First(&session).Error
	return err == nil
}
func IsRole(db *gorm.DB, r *http.Request, roleID int) bool {
	// Get the session cookie
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	// Find the session by its ID
	var session models.Session
	err = db.Where("session_id = ?", sessionCookie.Value).First(&session).Error
	if err != nil {
		return false
	}

	// Check the user's role
	var userRoleID int
	err = db.Table("users").Select("role_id").Where("id = ?", session.UserID).Row().Scan(&userRoleID)
	if err != nil {
		return false
	}

	// Return true if the user's role matches the provided roleID
	return userRoleID == roleID
}
func IsAdmin(db *gorm.DB, r *http.Request) bool {
	return IsRole(db, r, 4)
}
