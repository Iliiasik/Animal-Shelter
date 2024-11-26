package handlers

import (
	"Animals_Shelter/models"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type PageDataAnimals struct {
	LoggedIn        bool
	Animals         []AnimalWithDetails
	CurrentCategory string
}

func AnimalListPage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	loggedIn := false

	// Проверяем сессию
	session, err := r.Cookie("session")
	if err == nil && session.Value != "" {
		loggedIn = true
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

	// Получаем животных с деталями из базы данных с фильтрами
	animals, err := fetchAllAnimalsWithFiltration(db, species, breed, color, age, gender)
	if err != nil {
		http.Error(w, "Error fetching animals", http.StatusInternalServerError)
		return
	}

	// Создаем структуру данных для страницы
	data := PageDataAnimals{
		LoggedIn:        loggedIn,
		Animals:         animals,
		CurrentCategory: currentCategory,
	}

	// Загружаем HTML-шаблон и передаем данные
	tmpl := template.Must(template.ParseFiles("templates/animal_list.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func fetchAllAnimalsWithFiltration(db *sql.DB, species, breed, color, age, gender string) ([]AnimalWithDetails, error) {
	var animals []AnimalWithDetails

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

	// Массив аргументов для параметров запроса
	var args []interface{}

	// Добавляем условия для каждого фильтра, если он задан
	if species != "" {
		query += " AND animaltypes.type_name = $1"
		args = append(args, species)
	}
	if breed != "" {
		query += " AND animals.breed LIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+breed+"%")
	}
	if color != "" {
		query += " AND animals.color LIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+color+"%")
	}
	if age != "" {
		query += " AND animals.age = $" + strconv.Itoa(len(args)+1)
		args = append(args, age)
	}
	if gender != "" {
		query += " AND genders.name = $" + strconv.Itoa(len(args)+1)
		args = append(args, gender)
	}

	// Выполняем запрос с параметрами
	log.Printf("Query params - species: %s, breed: %s, color: %s, age: %s, gender: %s", species, breed, color, age, gender)

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Обрабатываем строки результата
	for rows.Next() {
		var animal AnimalWithDetails
		if err := rows.Scan(
			&animal.ID, &animal.Name, &animal.Species, &animal.Breed, &animal.Age,
			&animal.Gender, &animal.Status, &animal.ArrivalDate, &animal.Description,
			&animal.Location, &animal.Weight, &animal.Color, &animal.IsSterilized, &animal.HasPassport); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		// Получаем изображения для каждого животного
		images, err := fetchAnimalImages(db, animal.ID)
		if err != nil {
			log.Printf("Error fetching images for animal ID %d: %v", animal.ID, err)
			return nil, err
		}
		animal.Images = images

		// Добавляем животное в список
		animals = append(animals, animal)
	}

	// Проверяем на наличие ошибок после завершения цикла
	if err := rows.Err(); err != nil {
		log.Printf("Error after rows iteration: %v", err)
		return nil, err
	}

	return animals, nil
}

func AnimalInformation(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем ID животного из параметров запроса
	animalIDStr := r.URL.Query().Get("id")
	if animalIDStr == "" {
		http.Error(w, "Animal ID is required", http.StatusBadRequest)
		return
	}

	animalID, err := strconv.Atoi(animalIDStr)
	if err != nil {
		http.Error(w, "Invalid Animal ID", http.StatusBadRequest)
		return
	}

	// Создаём экземпляр структуры AnimalWithDetails
	var animal AnimalWithDetails

	// Выполняем SQL-запрос для получения информации о животном
	query := `
		SELECT animals.id, animals.name, animaltypes.type_name AS species, animals.breed, animals.age, 
		       genders.name AS gender, animalstatus.status_name AS status, animals.arrival_date, 
		       animals.description, animals.location, animals.weight, animals.color, 
		       animals.is_sterilized, animals.has_passport
		FROM animals
		JOIN animaltypes ON animals.species_id = animaltypes.id
		JOIN genders ON animals.gender_id = genders.id
		JOIN animalstatus ON animals.status_id = animalstatus.id
		WHERE animals.id = $1
	`
	err = db.QueryRow(query, animalID).Scan(
		&animal.ID,
		&animal.Name,
		&animal.Species,
		&animal.Breed,
		&animal.Age,
		&animal.Gender,
		&animal.Status,
		&animal.ArrivalDate,
		&animal.Description,
		&animal.Location,
		&animal.Weight,
		&animal.Color,
		&animal.IsSterilized,
		&animal.HasPassport,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Выполняем запрос для получения изображений животного
	query = `SELECT image_url FROM postimages WHERE animal_id = $1`
	rows, err := db.Query(query, animalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Заполняем список изображений
	for rows.Next() {
		var image models.PostImage
		if err := rows.Scan(&image.ImageURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		animal.Images = append(animal.Images, image)
	}

	// Рендеринг шаблона
	tmpl, err := template.ParseFiles("templates/animal_information.html")
	if err != nil {
		log.Printf("Error parsing template: %v\n", err)
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, animal); err != nil {
		log.Printf("Error executing template: %v\n", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
