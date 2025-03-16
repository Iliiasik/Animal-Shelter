package middleware

import (
	"Animals_Shelter/auth"
	"log"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
)

// RedirectIfLoggedIn перенаправляет пользователя на /profile, если он уже вошел
func RedirectIfLoggedIn(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.IsLoggedIn(db, r) {
			http.Redirect(w, r, "/profile?already_logged_in=true", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// LoggerMiddleware - middleware для логирования HTTP-запросов и ответов
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Создаем ResponseWriter обертку
		ww := &responseWriter{w, http.StatusOK}
		// Вызываем следующий обработчик
		next.ServeHTTP(ww, r)
		// Логируем запрос и ответ
		log.Printf("%s %s %d %s", r.Method, r.RequestURI, ww.status, time.Since(start))
	})
}

// responseWriter - обертка для http.ResponseWriter, чтобы захватывать статус код
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
