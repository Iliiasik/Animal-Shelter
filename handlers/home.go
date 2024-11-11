package handlers

import (
	"Animals_Shelter/models"
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

// PageData represents the data passed to the HTML templates
type PageData struct {
	LoggedIn        bool
	IsAdmin         bool
	Animals         []AnimalWithImages
	CurrentCategory string
}

// HomePage renders the home page
func HomePage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Проверяем статус входа пользователя
	loggedIn := false
	isAdmin := false
	session, err := r.Cookie("session")
	if err == nil && session.Value != "" {
		loggedIn = true

		// Check if the user is admin
		userID, err := getUserIDFromSession(db, session.Value)
		if err == nil {
			isAdmin, err = isUserAdmin(db, userID)
			if err != nil {
				http.Error(w, "Error checking admin status", http.StatusInternalServerError)
				return
			}
		}
	}

	// Получаем параметр species из запроса
	species := r.URL.Query().Get("species")

	// Устанавливаем CurrentCategory
	currentCategory := "all" // Значение по умолчанию
	if species != "" {
		currentCategory = species
	}

	// Fetch animals from the database
	animals, err := fetchAllAnimalsWithImages(db, species)
	if err != nil {
		http.Error(w, "Error fetching animals", http.StatusInternalServerError)
		return
	}

	data := PageData{
		LoggedIn:        loggedIn,
		IsAdmin:         isAdmin,
		Animals:         animals,
		CurrentCategory: currentCategory, // Устанавливаем CurrentCategory
	}
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	tmpl.Execute(w, data)
}

func AnimalListPage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Проверяем статус входа пользователя
	loggedIn := false
	isAdmin := false
	session, err := r.Cookie("session")
	if err == nil && session.Value != "" {
		loggedIn = true

		// Check if the user is admin
		userID, err := getUserIDFromSession(db, session.Value)
		if err == nil {
			isAdmin, err = isUserAdmin(db, userID)
			if err != nil {
				http.Error(w, "Error checking admin status", http.StatusInternalServerError)
				return
			}
		}
	}

	// Получаем параметры из запроса
	species := r.URL.Query().Get("species")
	breed := r.URL.Query().Get("breed")
	color := r.URL.Query().Get("color")
	age := r.URL.Query().Get("age")
	gender := r.URL.Query().Get("gender")

	// Устанавливаем CurrentCategory
	currentCategory := "all" // Значение по умолчанию
	if species != "" {
		currentCategory = species
	}

	// Fetch animals from the database с использованием фильтров
	animals, err := fetchAllAnimals(db, species, breed, color, age, gender)
	if err != nil {
		http.Error(w, "Error fetching animals", http.StatusInternalServerError)
		return
	}

	data := PageData{
		LoggedIn:        loggedIn,
		IsAdmin:         isAdmin,
		Animals:         animals,
		CurrentCategory: currentCategory, // Устанавливаем CurrentCategory
	}
	tmpl := template.Must(template.ParseFiles("templates/animal_list.html"))
	tmpl.Execute(w, data)
}

// Helper function to get user ID from session
func getUserIDFromSession(db *sql.DB, sessionID string) (int, error) {
	var userID int
	now := time.Now()
	query := `
		SELECT user_id FROM sessions
		WHERE session_id = $1 AND expires_at > $2
	`
	err := db.QueryRow(query, sessionID, now).Scan(&userID)
	return userID, err
}

// Helper function to check if the user is an admin
func isUserAdmin(db *sql.DB, userID int) (bool, error) {
	var isAdmin bool
	query := `SELECT is_admin FROM Users WHERE id = $1`
	err := db.QueryRow(query, userID).Scan(&isAdmin)
	return isAdmin, err
}
func fetchAllAnimals(db *sql.DB, species, breed, color, age, gender string) ([]AnimalWithImages, error) {
	var query string
	var args []interface{}

	// Начинаем с базового запроса
	query = `
		SELECT id
		FROM Animals
		WHERE 1=1
	`

	// Добавляем условия в запрос в зависимости от переданных параметров
	if species != "" {
		query += " AND species = $1"
		args = append(args, species)
	}
	if breed != "" {
		query += " AND breed LIKE $" + strconv.Itoa(len(args)+1) // Используем индексацию для параметров
		args = append(args, "%"+breed+"%")
	}
	if color != "" {
		query += " AND color LIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+color+"%")
	}
	if age != "" {
		query += " AND age = $" + strconv.Itoa(len(args)+1)
		args = append(args, age)
	}
	if gender != "" {
		query += " AND gender = $" + strconv.Itoa(len(args)+1)
		args = append(args, gender)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var animals []AnimalWithImages
	for rows.Next() {
		var animalID int
		if err := rows.Scan(&animalID); err != nil {
			return nil, err
		}

		animal, err := fetchAnimalInformation(db, animalID)
		if err != nil {
			return nil, err
		}

		animals = append(animals, animal)
	}

	return animals, nil
}

func fetchAllAnimalsWithImages(db *sql.DB, species string) ([]AnimalWithImages, error) {
	var query string
	var args []interface{}

	if species != "" {
		query = `
			SELECT id
			FROM Animals
			WHERE species = $1
		`
		args = append(args, species)
	} else {
		query = `
			SELECT id
			FROM Animals
		`
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var animals []AnimalWithImages
	for rows.Next() {
		var animalID int
		if err := rows.Scan(&animalID); err != nil {
			return nil, err
		}

		animal, err := fetchAnimalInformation(db, animalID)
		if err != nil {
			return nil, err
		}

		animals = append(animals, animal)
	}

	return animals, nil
}

func fetchAnimalInformation(db *sql.DB, animalID int) (AnimalWithImages, error) {
	// Fetch animal information
	var animal AnimalWithImages
	query := `
		SELECT id, name, species, breed, age, gender, status_id, arrival_date, description
		FROM Animals
		WHERE id = $1
	`
	err := db.QueryRow(query, animalID).Scan(&animal.ID, &animal.Name, &animal.Species, &animal.Breed, &animal.Age, &animal.Gender, &animal.StatusID, &animal.ArrivalDate, &animal.Description)
	if err != nil {
		return animal, err
	}

	// Fetch animal images
	imageQuery := `
		SELECT id, animal_id, image_url
		FROM PostImages
		WHERE animal_id = $1
	`
	rows, err := db.Query(imageQuery, animalID)
	if err != nil {
		return animal, err
	}
	defer rows.Close()

	for rows.Next() {
		var image models.PostImage
		if err := rows.Scan(&image.ID, &image.AnimalID, &image.ImageURL); err != nil {
			return animal, err
		}
		animal.Images = append(animal.Images, image)
	}

	return animal, nil
}
func TermsOfServicePage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/user_agreement.html")
}
