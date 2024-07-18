package main

import (
	"log"
	"net/http"

	"Animals_Shelter/db"
	"Animals_Shelter/handlers"

	_ "github.com/lib/pq"
)

func main() {
	// Подключение к базе данных
	database := db.ConnectDB()
	defer database.Close()

	// Настройка маршрутов
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HomePage(database, w, r)
	})
	http.HandleFunc("/animals", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowAddAnimalForm(w, r)
		case "POST":
			handlers.AddAnimal(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/animal_information", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.AnimalInformation(database, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/medical_records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowAddMedicalRecordForm(database, w, r)
		case "POST":
			handlers.AddMedicalRecord(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowRegisterForm(w, r)
		case "POST":
			handlers.Register(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.ShowLoginForm(w, r)
		case "POST":
			handlers.Login(database, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.Logout(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/confirm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.ConfirmEmail(database, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
