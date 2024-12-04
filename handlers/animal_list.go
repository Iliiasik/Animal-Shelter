package handlers

import (
	"Animals_Shelter/models"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// Для информации о животных

type AnimalWithDetails struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	Species      string             `json:"species"`
	Breed        string             `json:"breed"`
	AgeYears     int                `json:"age_years"`  // Год
	AgeMonths    int                `json:"age_months"` // Месяцы
	Gender       string             `json:"gender"`
	Status       string             `json:"status"`
	ArrivalDate  string             `json:"arrival_date"`
	Description  string             `json:"description"`
	Location     string             `json:"location"`
	Weight       float64            `json:"weight"` // Тип изменен на float
	Color        string             `json:"color"`
	IsSterilized bool               `json:"is_sterilized"`
	HasPassport  bool               `json:"has_passport"`
	Views        int                `json:"views"`
	Images       []models.PostImage `json:"images"`
	UserDetails  struct {
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		PhoneNumber  string `json:"phone_number"`
		Email        string `json:"email"`
		ProfileImage string `json:"profile_image"` // Добавляем профильное изображение
		Username     string `json:"username"`
		BgImage      string `json:"profile_bg_image"`
	} `json:"user_details"`
}

// Для листа животных

type PageDataAnimals struct {
	LoggedIn        bool
	Animals         []AnimalSummary
	CurrentCategory string
}
type AnimalSummary struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Species   string `json:"species"`
	Breed     string `json:"breed"`
	Color     string `json:"color"`
	AgeYears  int    `json:"age_years"`  // Число лет
	AgeMonths int    `json:"age_months"` // Число месяцев
	Gender    string `json:"gender"`
	ImageURL  string `json:"image_url"`
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
	ageYears := r.URL.Query().Get("age_years")
	ageMonths := r.URL.Query().Get("age_months")
	gender := r.URL.Query().Get("gender")

	// Устанавливаем CurrentCategory
	currentCategory := "all" // Значение по умолчанию
	if species != "" {
		currentCategory = species
	}

	// Получаем животных с деталями из базы данных с фильтрами
	animals, err := fetchAnimalsWithFilters(db, species, breed, color, ageYears, ageMonths, gender)
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

func fetchAnimalsWithFilters(db *sql.DB, species, breed, color, ageYearsStr, ageMonthsStr, gender string) ([]AnimalSummary, error) {
	var animals []AnimalSummary

	query := `
        SELECT animals.id, animals.name, animaltypes.type_name AS species, animals.breed, 
               animals.color, animalages.years, animalages.months, genders.name AS gender,
               (SELECT image_url FROM postimages WHERE animal_id = animals.id LIMIT 1) AS image
        FROM animals
        JOIN animaltypes ON animals.species_id = animaltypes.id
        JOIN genders ON animals.gender_id = genders.id
        LEFT JOIN animalages ON animals.id = animalages.animal_id
        WHERE 1=1
    `

	var args []interface{}

	// Преобразование строк в числа для возраста
	var ageYears, ageMonths int
	var err error
	if ageYearsStr != "" {
		ageYears, err = strconv.Atoi(ageYearsStr)
		if err != nil {
			log.Printf("Invalid ageYears: %v", err)
			return nil, err
		}
	}
	if ageMonthsStr != "" {
		ageMonths, err = strconv.Atoi(ageMonthsStr)
		if err != nil {
			log.Printf("Invalid ageMonths: %v", err)
			return nil, err
		}
	}

	// Фильтрация по species
	if species != "" {
		query += " AND animaltypes.type_name = $" + strconv.Itoa(len(args)+1)
		args = append(args, species)
	}

	// Фильтрация по breed
	if breed != "" {
		query += " AND animals.breed LIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+breed+"%")
	}

	// Фильтрация по color
	if color != "" {
		query += " AND animals.color LIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%"+color+"%")
	}

	// Фильтрация по возрасту (если ageYears и ageMonths предоставлены)
	if ageYearsStr != "" || ageMonthsStr != "" {
		// Если указаны оба значения возраста
		if ageYearsStr != "" && ageMonthsStr != "" {
			query += " AND animalages.years = $" + strconv.Itoa(len(args)+1) + " AND animalages.months = $" + strconv.Itoa(len(args)+2)
			args = append(args, ageYears, ageMonths)
		} else if ageYearsStr != "" {
			// Если указан только возраст в годах
			query += " AND animalages.years = $" + strconv.Itoa(len(args)+1)
			args = append(args, ageYears)
		} else if ageMonthsStr != "" {
			// Если указан только возраст в месяцах
			query += " AND animalages.months = $" + strconv.Itoa(len(args)+1)
			args = append(args, ageMonths)
		}
	}

	// Фильтрация по gender
	if gender != "" {
		query += " AND genders.name = $" + strconv.Itoa(len(args)+1)
		args = append(args, gender)
	}

	// Выполняем запрос
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Обрабатываем строки результата
	for rows.Next() {
		var animal AnimalSummary
		if err := rows.Scan(
			&animal.ID, &animal.Name, &animal.Species, &animal.Breed,
			&animal.Color, &animal.AgeYears, &animal.AgeMonths, &animal.Gender, &animal.ImageURL); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		animals = append(animals, animal)
	}

	// Проверяем ошибки после итерации
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
		SELECT animals.id, animals.name, animaltypes.type_name AS species, animals.breed, 
		       genders.name AS gender, animalstatus.status_name AS status, animals.arrival_date, 
		       animals.description, animals.location, animals.weight, animals.color, 
		       animals.is_sterilized, animals.has_passport, animals.views
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
		&animal.Gender,
		&animal.Status,
		&animal.ArrivalDate,
		&animal.Description,
		&animal.Location,
		&animal.Weight,
		&animal.Color,
		&animal.IsSterilized,
		&animal.HasPassport,
		&animal.Views,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Запрос для получения возраста животного (из таблицы animalages)
	ageQuery := `SELECT years, months FROM animalages WHERE animal_id = $1`
	err = db.QueryRow(ageQuery, animalID).Scan(
		&animal.AgeYears,
		&animal.AgeMonths,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если возраст не найден, то оставляем значения по умолчанию (0)
			animal.AgeYears = 0
			animal.AgeMonths = 0
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

	// Выполняем запрос для получения информации о владельце и его профиле
	userQuery := `
		SELECT u.email, u.username, ud.first_name, ud.last_name, ud.phone_number, ui.profile_image, ui.profile_bg_image
		FROM users u
		JOIN user_details ud ON u.id = ud.user_id
		JOIN user_images ui ON u.id = ui.user_id
		WHERE u.id = (
			SELECT user_id FROM animals WHERE id = $1
		)
	`
	err = db.QueryRow(userQuery, animalID).Scan(
		&animal.UserDetails.Email,
		&animal.UserDetails.Username,
		&animal.UserDetails.FirstName,
		&animal.UserDetails.LastName,
		&animal.UserDetails.PhoneNumber,
		&animal.UserDetails.ProfileImage,
		&animal.UserDetails.BgImage,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No user details found for the animal.")
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

func IncrementViews(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	query := `UPDATE animals SET views = views + 1 WHERE id = $1`
	_, err = db.Exec(query, animalID)
	if err != nil {
		http.Error(w, "Failed to increment views", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
