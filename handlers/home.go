package handlers

import (
	"Animals_Shelter/models"
	"database/sql"
	"html/template"
	"net/http"
)

// PageData represents the data passed to the HTML templates
type PageData struct {
	LoggedIn bool
	Animals  []AnimalWithImages
}

// HomePage renders the home page
func HomePage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Проверяем статус входа пользователя
	loggedIn := false
	session, err := r.Cookie("session")
	if err == nil && session.Value != "" {
		loggedIn = true
	}

	// Получаем параметр species из запроса
	species := r.URL.Query().Get("species")

	// Fetch animals from the database
	animals, err := fetchAllAnimalsWithImages(db, species)
	if err != nil {
		http.Error(w, "Error fetching animals", http.StatusInternalServerError)
		return
	}

	data := PageData{
		LoggedIn: loggedIn,
		Animals:  animals,
	}
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	tmpl.Execute(w, data)
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
