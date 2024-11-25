package handlers

import (
	"Animals_Shelter/models"
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

// PageData represents the data passed to the HTML templates
type PageData struct {
	LoggedIn bool
	Animals  []AnimalWithDetails
}

// AnimalWithDetails represents an animal with its associated details
type AnimalWithDetails struct {
	ID           int
	Name         string
	Species      string
	Breed        string
	Age          int
	Gender       string
	Status       string
	ArrivalDate  string
	Description  string
	Location     string
	Weight       int
	Color        string
	IsSterilized bool
	HasPassport  bool
	Images       []models.PostImage
}

// HomePage handles rendering the homepage
func HomePage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	loggedIn := false

	// Проверяем сессию
	session, err := r.Cookie("session")
	if err == nil && session.Value != "" {
		loggedIn = true
	}

	// Получаем животных с деталями
	animals, err := fetchAllAnimalsWithDetails(db)
	if err != nil {
		http.Error(w, "Error fetching animals", http.StatusInternalServerError)
		return
	}

	data := PageData{
		LoggedIn: loggedIn,
		Animals:  animals,
	}
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		return
	}
}

func fetchAllAnimalsWithDetails(db *sql.DB) ([]AnimalWithDetails, error) {
	var animals []AnimalWithDetails

	// SQL-запрос для извлечения всех животных с их деталями
	query := `
		SELECT animals.id, animals.name, animaltypes.type_name AS species, animals.breed, animals.age, 
			genders.name AS gender, animalstatus.status_name AS status, animals.arrival_date, 
			animals.description, animals.location, animals.weight, animals.color, 
			animals.is_sterilized, animals.has_passport
		FROM animals
		JOIN animaltypes ON animals.species_id = animaltypes.id
		JOIN genders ON animals.gender_id = genders.id
		JOIN animalstatus ON animals.status_id = animalstatus.id
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обрабатываем все строки результата
	for rows.Next() {
		var animal AnimalWithDetails
		if err := rows.Scan(
			&animal.ID, &animal.Name, &animal.Species, &animal.Breed, &animal.Age,
			&animal.Gender, &animal.Status, &animal.ArrivalDate, &animal.Description,
			&animal.Location, &animal.Weight, &animal.Color, &animal.IsSterilized, &animal.HasPassport); err != nil {
			return nil, err
		}

		// Получаем изображения для каждого животного
		images, err := fetchAnimalImages(db, animal.ID)
		if err != nil {
			return nil, err
		}
		animal.Images = images

		// Добавляем животное в список
		animals = append(animals, animal)
	}

	// Проверяем на наличие ошибок после завершения цикла
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return animals, nil
}

func fetchAnimalImages(db *sql.DB, animalID int) ([]models.PostImage, error) {
	var images []models.PostImage

	// Подготовка запроса для получения изображений
	imageQuery := `
		SELECT id, animal_id, image_url
		FROM PostImages
		WHERE animal_id = $1
	`
	rows, err := db.Query(imageQuery, animalID)
	if err != nil {
		// Логируем ошибку выполнения запроса
		log.Printf("Error executing query: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	// Обрабатываем все строки результата
	for rows.Next() {
		var image models.PostImage
		if err := rows.Scan(&image.ID, &image.AnimalID, &image.ImageURL); err != nil {
			// Логируем ошибку, если не удалось считать строку
			log.Printf("Error scanning row: %v\n", err)
			return nil, err
		}
		images = append(images, image)
	}

	// Проверка на ошибки после выполнения цикла
	if err := rows.Err(); err != nil {
		// Логируем ошибку
		log.Printf("Error iterating over rows: %v\n", err)
		return nil, err
	}

	return images, nil
}

// TermsOfServicePage serves the terms of service page
func TermsOfServicePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/user_agreement.html")
}
