package main

import (
	"Animals_Shelter/admin"
	"Animals_Shelter/auth"
	"Animals_Shelter/db"
	"Animals_Shelter/handlers"
	"Animals_Shelter/middleware"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

func main() {
	gormDB := db.ConnectDB()
	Admin := admin.InitAdmin(gormDB)

	sqlDB := gormDB.DB()
	if sqlDB == nil {
		log.Fatalf("Failed to get *sql.DB from GORM")
	}

	defer func(sqlDB *sql.DB) {
		err := sqlDB.Close()
		if err != nil {

		}
	}(sqlDB)

	mux := http.NewServeMux()
	adminHandler := middleware.AdminAuthMiddleware(gormDB, Admin.NewServeMux("/admin"), auth.IsLoggedIn, auth.IsAdmin)
	mux.Handle("/admin/", adminHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HomePage(gormDB, w, r)
	})

	mux.HandleFunc("/animal_list", func(w http.ResponseWriter, r *http.Request) {
		handlers.AnimalListPage(sqlDB, w, r)
	})

	mux.HandleFunc("/animals/delete", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteAnimal(gormDB, w, r)
	})
	mux.HandleFunc("/add-animal", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddAnimal(gormDB, w, r)
	})

	mux.HandleFunc("/animal_information", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.AnimalInformation(sqlDB, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/adopt", func(w http.ResponseWriter, r *http.Request) {
		handlers.RegisterAdoption(sqlDB, w, r)
	})
	mux.HandleFunc("/update_adoption_status", func(w http.ResponseWriter, r *http.Request) {
		handlers.AcceptAdoption(sqlDB, w, r)
	})
	mux.HandleFunc("/delete_adoption", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeclineAdoption(sqlDB, w, r)
	})
	mux.HandleFunc("/deal_canceled", func(w http.ResponseWriter, r *http.Request) {
		handlers.DealCanceled(sqlDB, w, r)
	})
	mux.HandleFunc("/transfer_animal", func(w http.ResponseWriter, r *http.Request) {
		handlers.TransferAnimal(sqlDB, w, r)
	})
	mux.HandleFunc("/increment_views", func(w http.ResponseWriter, r *http.Request) {
		handlers.IncrementViews(sqlDB, w, r)
	})
	mux.HandleFunc("/update-rating", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateRating(sqlDB, w, r)
	})
	mux.HandleFunc("/register", middleware.RedirectIfLoggedIn(gormDB, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowRegisterForm(w)
		case "POST":
			handlers.Register(gormDB, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/login", middleware.RedirectIfLoggedIn(gormDB, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowLoginForm(w)
		case "POST":
			handlers.Login(gormDB, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

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
	mux.HandleFunc("/save-profile", func(w http.ResponseWriter, r *http.Request) {
		handlers.SaveProfile(gormDB, w, r) // Маршрут для сохранения профиля
	})
	mux.HandleFunc("/save-visibility-settings", func(w http.ResponseWriter, r *http.Request) {
		handlers.SaveVisibilitySettings(gormDB, w, r)
	})
	mux.HandleFunc("/profile/{username}", func(w http.ResponseWriter, r *http.Request) {
		re := regexp.MustCompile(`/profile/([a-zA-Z0-9_]+)`)
		match := re.FindStringSubmatch(r.URL.Path)
		if len(match) > 1 {
			username := match[1]
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
	mux.HandleFunc("/toggle_like", func(w http.ResponseWriter, r *http.Request) {
		handlers.ToggleLike(sqlDB, w, r)
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

	loggedMux := middleware.LoggerMiddleware(mux)
	port := 8080
	address := fmt.Sprintf("http://34.16.104.66:%d", port)
	fmt.Printf("Starting server on %s\n", address)

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}
