package main

import (
	"Animals_Shelter/admin"
	"Animals_Shelter/admin/auth"
	"Animals_Shelter/admin/db_old"
	"Animals_Shelter/admin/middleware"
	"Animals_Shelter/db"
	"Animals_Shelter/handlers"
	"Animals_Shelter/storage"
	"fmt"
	_ "github.com/minio/minio-go/v7"
	"log"
	"net/http"
	"regexp"
	"time"
)

func main() {
	// Подключение к базе данных через GORM
	gormDB := db.ConnectDB()

	oldGormDB := db_old.ConnectOldDB()
	// Initialize QOR Admin
	// Инициализация админки
	Admin := admin.InitAdmin(oldGormDB)

	// Получаем *sql.DB из *gorm.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Failed to get *sql.DB from GORM: %v", err)
	}

	defer sqlDB.Close()

	// Создаем новый маршрутизатор
	mux := http.NewServeMux()

	// Инициализация MinIO клиента
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"

	minioService, err := storage.InitMinioClient(endpoint, accessKeyID, secretAccessKey)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// Оборачиваем админку в Middleware
	adminHandler := middleware.AdminAuthMiddleware(gormDB, Admin.NewServeMux("/admin"), auth.IsLoggedIn, auth.IsAdmin)
	mux.Handle("/admin/", adminHandler)

	// Настройка маршрутов с использованием *sql.DB
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HomePage(sqlDB, w, r)
	})

	mux.HandleFunc("/animal_list", func(w http.ResponseWriter, r *http.Request) {
		handlers.AnimalListPage(sqlDB, w, r)
	})

	mux.HandleFunc("/animals", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowAddAnimalForm(w, r)
		case "POST":
			handlers.AddAnimal(sqlDB, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/animal_information", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.AnimalInformation(sqlDB, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/medical_records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowAddMedicalRecordForm(sqlDB, w, r)
		case "POST":
			handlers.AddMedicalRecord(sqlDB, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowRegisterForm(w)
		case "POST":
			handlers.Register(gormDB, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowLoginForm(w)
		case "POST":
			handlers.Login(gormDB, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.Logout(gormDB, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/confirm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.ConfirmEmail(gormDB, w, r)
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
	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.ShowProfile(gormDB, w, r) // Маршрут для профиля
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/edit-profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.RenderEditTemplate(gormDB, w, r) // Маршрут для редактирования профиля
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/save-profile", func(w http.ResponseWriter, r *http.Request) {
		handlers.SaveProfile(gormDB, minioService, w, r)
	})

	mux.HandleFunc("/save-visibility-settings", func(w http.ResponseWriter, r *http.Request) {
		handlers.SaveVisibilitySettings(gormDB, w, r)
	})
	mux.HandleFunc("/profile/{username}", func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем username из URL с помощью регулярного выражения
		re := regexp.MustCompile(`/profile/([a-zA-Z0-9_]+)`) // Можно настроить регулярку под ваши требования для username
		match := re.FindStringSubmatch(r.URL.Path)

		// Если путь соответствует, извлекаем username
		if len(match) > 1 {
			username := match[1]
			// Подключаемся к базе данных (или используем уже подключенную)
			// Пример обработки маршрута с username:
			handlers.ViewProfile(gormDB, w, username)
		} else {
			http.Error(w, "Invalid URL", http.StatusNotFound)
		}
	})
	mux.HandleFunc("/feedback", func(w http.ResponseWriter, r *http.Request) {
		handlers.LoadFeedbackPage(w, r) // Обработчик для отображения страницы отзывов
	})
	mux.HandleFunc("/feedback-save", handlers.SaveFeedback(gormDB)) // Обработчик для сохранения отзыва
	mux.HandleFunc("/forum", func(w http.ResponseWriter, r *http.Request) {
		handlers.ShowForum(sqlDB, w, r)
	})
	mux.HandleFunc("/create_topic", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateTopic(sqlDB, w, r)
	})
	mux.HandleFunc("/create_post", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreatePost(sqlDB, w, r)
	})
	mux.HandleFunc("/topic", func(w http.ResponseWriter, r *http.Request) {
		handlers.ShowTopic(sqlDB, w, r)
	})
	mux.HandleFunc("/delete-topics", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteTopics(gormDB, w, r)
	})

	mux.HandleFunc("/terms-of-service", func(w http.ResponseWriter, r *http.Request) {
		handlers.TermsOfServicePage(w, r)
	})

	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	mux.Handle("/system_images/", http.StripPrefix("/system_images/", http.FileServer(http.Dir("system_images"))))

	// Оборачиваем маршрутизатор в middleware логирования
	loggedMux := LoggerMiddleware(mux)

	port := 8080
	address := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf("Starting server on %s\n", address)

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
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
