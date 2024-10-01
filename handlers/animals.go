package handlers

import (
	"Animals_Shelter/models"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type AnimalWithImages struct {
	models.Animal
	Images      []models.PostImage
	Status      string
	SpeciesName string
}

// ShowAddAnimalForm displays the form to add a new animal
func ShowAddAnimalForm(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/add_animal.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// AddAnimal handles the submission of the add animal form
func AddAnimal(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	var animal models.Animal
	animal.Name = r.FormValue("name")

	// Получаем или создаем запись о виде (species)
	speciesName := r.FormValue("species")
	var speciesID int
	err := db.QueryRow("SELECT id FROM animaltypes WHERE type_name = $1", speciesName).Scan(&speciesID)
	if err == sql.ErrNoRows {
		// Если вид отсутствует, добавляем его
		err = db.QueryRow("INSERT INTO animaltypes (type_name) VALUES ($1) RETURNING id", speciesName).Scan(&speciesID)
		if err != nil {
			http.Error(w, "Error inserting species", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Error fetching species", http.StatusInternalServerError)
		return
	}
	animal.Species = speciesID

	// Получаем или создаем запись о статусе
	statusName := r.FormValue("status_id")
	var statusID int
	err = db.QueryRow("SELECT id FROM animalstatus WHERE status_name = $1", statusName).Scan(&statusID)
	if err == sql.ErrNoRows {
		// Если статус отсутствует, добавляем его
		err = db.QueryRow("INSERT INTO animalstatus (status_name) VALUES ($1) RETURNING id", statusName).Scan(&statusID)
		if err != nil {
			http.Error(w, "Error inserting status", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Error fetching status", http.StatusInternalServerError)
		return
	}
	animal.StatusID = statusID

	// Остальные данные животного
	animal.Breed = r.FormValue("breed")
	animal.Age, _ = strconv.Atoi(r.FormValue("age"))
	animal.Gender = r.FormValue("gender")
	animal.ArrivalDate = r.FormValue("arrival_date")
	animal.Description = r.FormValue("description")

	// Создаем директорию "uploads", если она не существует
	uploadDir := "uploads"

	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err = os.Mkdir(uploadDir, os.ModePerm)
		if err != nil {
			http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
			return
		}
	}

	var filePaths []string

	// Обрабатываем загрузку нескольких изображений
	files := r.MultipartForm.File["images"]
	if len(files) > 4 {
		http.Error(w, "You can upload a maximum of 4 images", http.StatusBadRequest)
		return
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Генерация уникального имени файла
		fileExt := path.Ext(fileHeader.Filename)
		fileName := strconv.FormatInt(time.Now().UnixNano(), 10) + fileExt
		filePath := path.Join(uploadDir, fileName)

		// Сохранение файла на сервере
		outFile, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		_, err = file.Seek(0, 0)
		if err != nil {
			http.Error(w, "Error seeking file", http.StatusInternalServerError)
			return
		}

		_, err = outFile.ReadFrom(file)
		if err != nil {
			http.Error(w, "Error writing file", http.StatusInternalServerError)
			return
		}

		// Конвертируем обратные слеши в прямые для правильного пути файла
		filePath = filepath.ToSlash(filePath)
		filePaths = append(filePaths, filePath)
	}

	// Вставка данных о животном в базу данных
	query := `
        INSERT INTO animals (name, species, breed, age, gender, status_id, arrival_date, description)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `
	var animalID int
	err = db.QueryRow(query, animal.Name, animal.Species, animal.Breed, animal.Age, animal.Gender, animal.StatusID, animal.ArrivalDate, animal.Description).Scan(&animalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Вставка данных изображений в базу данных
	imageQuery := `
        INSERT INTO postimages (animal_id, image_url)
        VALUES ($1, $2)
    `
	for _, filePath := range filePaths {
		_, err = db.Exec(imageQuery, animalID, filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Перенаправление на главную страницу после успешной обработки
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ShowAddMedicalRecordForm displays the form to add a new medical record
func ShowAddMedicalRecordForm(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Fetch animals from the database to populate the dropdown
	rows, err := db.Query("SELECT id, name FROM Animals")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var animals []models.Animal
	for rows.Next() {
		var animal models.Animal
		if err := rows.Scan(&animal.ID, &animal.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		animals = append(animals, animal)
	}

	// Pass animals to the template
	tmpl, err := template.ParseFiles("templates/add_medical_record.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, struct{ Animals []models.Animal }{Animals: animals})
}

// AddMedicalRecord handles the submission of the add medical record form
func AddMedicalRecord(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	animalID, err := strconv.Atoi(r.FormValue("animal_id"))
	if err != nil {
		http.Error(w, "Invalid animal ID", http.StatusBadRequest)
		return
	}
	checkupDate := r.FormValue("checkup_date")
	notes := r.FormValue("notes")
	vetName := r.FormValue("vet_name")

	query := `
		INSERT INTO MedicalRecords (animal_id, checkup_date, notes, vet_name)
		VALUES ($1, $2, $3, $4)
	`
	_, err = db.Exec(query, animalID, checkupDate, notes, vetName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// AnimalInformation handles the request to view a specific animal's information
func AnimalInformation(db *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	// Fetch animal information
	var animal AnimalWithImages
	query := `
        SELECT a.id, a.name, a.species, t.type_name, a.breed, a.age, a.gender, a.status_id, s.status_name, a.arrival_date, a.description
        FROM animals a
        JOIN animalstatus s ON a.status_id = s.id
        JOIN animaltypes t ON a.species = t.id
        WHERE a.id = $1
    `
	err = db.QueryRow(query, animalID).Scan(&animal.ID, &animal.Name, &animal.Species, &animal.SpeciesName, &animal.Breed, &animal.Age, &animal.Gender, &animal.StatusID, &animal.Status, &animal.ArrivalDate, &animal.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Animal not found", http.StatusNotFound)
		} else {
			log.Printf("Error fetching animal information: %v\n", err)
			http.Error(w, "Failed to fetch animal information", http.StatusInternalServerError)
		}
		return
	}

	// Fetch animal images
	imageQuery := `
        SELECT id, animal_id, image_url
        FROM postimages
        WHERE animal_id = $1
    `
	rows, err := db.Query(imageQuery, animalID)
	if err != nil {
		log.Printf("Error fetching animal images: %v\n", err)
		http.Error(w, "Failed to fetch animal images", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var image models.PostImage
		if err := rows.Scan(&image.ID, &image.AnimalID, &image.ImageURL); err != nil {
			log.Printf("Error scanning image row: %v\n", err)
			http.Error(w, "Failed to scan image row", http.StatusInternalServerError)
			return
		}
		animal.Images = append(animal.Images, image)
	}

	// Render template with animal data
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
