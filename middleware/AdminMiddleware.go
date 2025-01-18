package middleware

import (
	"github.com/jinzhu/gorm"
	"net/http"
)

func AdminAuthMiddleware(db *gorm.DB, next http.Handler, isLoggedInCheck func(db *gorm.DB, r *http.Request) bool, isAdminCheck func(db *gorm.DB, r *http.Request) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is logged in
		if !isLoggedInCheck(db, r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Check if the user is an admin
		if !isAdminCheck(db, r) {
			http.Error(w, "Page Not Found", http.StatusNotFound)
			return
		}

		// Proceed with the request if both checks pass
		next.ServeHTTP(w, r)
	})
}
