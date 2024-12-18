package middleware

import (
	"gorm.io/gorm"
	"net/http"
)

// AdminAuthMiddleware проверяет, является ли пользователь администратором
func AdminAuthMiddleware(db *gorm.DB, next http.Handler, isLoggedInCheck func(db *gorm.DB, r *http.Request) bool, isAdminCheck func(db *gorm.DB, r *http.Request) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLoggedInCheck(db, r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if !isAdminCheck(db, r) {
			http.Error(w, "Page Not Found", http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
