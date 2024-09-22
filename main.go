package main

import (
	"log"
	"net/http"
	"time"

	"Animals_Shelter/db"
	"Animals_Shelter/handlers"

	_ "github.com/lib/pq"
)

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

func main() {
	// Подключение к базе данных
	database := db.ConnectDB()
	defer database.Close()

	// Создаем новый маршрутизатор
	mux := http.NewServeMux()

	// Настройка маршрутов
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HomePage(database, w, r)
	})

	mux.HandleFunc("/animals", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowAddAnimalForm(w, r)
		case "POST":
			handlers.AddAnimal(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/animal_information", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.AnimalInformation(database, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/medical_records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowAddMedicalRecordForm(database, w, r)
		case "POST":
			handlers.AddMedicalRecord(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowRegisterForm(w, r)
		case "POST":
			handlers.Register(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowLoginForm(w, r)
		case "POST":
			handlers.Login(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.Logout(database, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/confirm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.ConfirmEmail(database, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/registration_success", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "templates/registration_success.html")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		handlers.AdminPanel(database, w, r)
	})
	mux.HandleFunc("/admin/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			handlers.DeleteRecord(database, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Оборачиваем маршрутизатор в middleware логирования
	loggedMux := LoggerMiddleware(mux)

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}
