package handlers

import (
	"Animals_Shelter/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AddAnimal handles the submission of the add animal form
func AddAnimal(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	log.Println("Start AddAnimal handler")

	// Устанавливаем заголовок ответа как JSON
	w.Header().Set("Content-Type", "application/json")

	// Парсим форму с ограничением 10 МБ
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Println("Error parsing form data:", err)
		respondWithJSON(w, http.StatusBadRequest, "error", "Error parsing form data")
		return
	}

	// Обработка сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		log.Println("Session not found:", err)
		respondWithJSON(w, http.StatusUnauthorized, "error", "Session not found")
		return
	}

	var session Session
	if err := db.Where("session_id = ?", cookie.Value).First(&session).Error; err != nil {
		log.Println("Error fetching user ID from session:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error fetching user ID from session")
		return
	}
	log.Printf("UserID: %d\n", session.UserID)

	// Проверяем, сколько объявлений уже создал пользователь
	var animalCount int64
	if err := db.Model(&models.Animal{}).Where("user_id = ?", session.UserID).Count(&animalCount).Error; err != nil {
		log.Println("Error counting user animals:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error checking user advertisement limit")
		return
	}

	// Проверка ограничения на количество объявлений
	if animalCount >= 10 {
		log.Printf("User %d has reached the advertisement limit\n", session.UserID)
		respondWithJSON(w, http.StatusForbidden, "error", "Advertisement limit reached. You can create up to 10 animals.")
		return
	}

	// Создаем новый объект Animal
	animal := models.Animal{
		Name:        r.FormValue("name"),
		Breed:       r.FormValue("breed"),
		Description: r.FormValue("description"),
		Location:    r.FormValue("location"),
		Color:       r.FormValue("color"),

		UserID:       session.UserID,
		IsSterilized: parseBool(r.FormValue("is_sterilized")),
		HasPassport:  parseBool(r.FormValue("has_passport")),
	}

	weight, err := strconv.ParseFloat(r.FormValue("weight"), 64)
	if err != nil {
		weight = 0.0
	}
	animal.Weight = weight
	log.Printf("Animal data: %+v\n", animal)

	// Обработка связанных данных: species, status, gender
	if err := processRelation(db, &animal.Species, "name", r.FormValue("species")); err != nil {
		log.Println("Error processing species:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error processing species")
		return
	}
	log.Printf("Species: %+v\n", animal.Species)

	var status models.AnimalStatus
	if err := db.Where("id = ?", 1).First(&status).Error; err != nil {
		log.Println("Error fetching status with ID 4:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error fetching status")
		return
	}
	animal.Status = status

	log.Printf("Status: %+v\n", animal.Status)

	if err := processRelation(db, &animal.Gender, "name", r.FormValue("gender")); err != nil {
		log.Println("Error processing gender:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error processing gender")
		return
	}
	log.Printf("Gender: %+v\n", animal.Gender)
	// Проверка изображений перед сохранением
	if fileExt, err := validateAnimalImages(r); err != nil {
		log.Println("Error validating images:", err)
		respondWithJSON(w, http.StatusBadRequest, "error", fmt.Sprintf("Invalid image extension: %s", fileExt))
		return
	}
	// Вставка данных животного
	if err := db.Create(&animal).Error; err != nil {
		log.Println("Error inserting animal:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error inserting animal")
		return
	}

	// Обработка возраста животного
	if err := saveAnimalAge(db, &animal, r.FormValue("age_years"), r.FormValue("age_months")); err != nil {
		log.Println("Error saving animal age:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error saving animal age")
		return
	}
	// Обработка изображений
	if fileExt, err := processAnimalImages(db, r, uint(animal.ID)); err != nil {
		log.Println("Error processing images:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", fmt.Sprintf("Error processing images. Invalid extension: %s", fileExt))
		return
	}

	log.Println("Animal inserted successfully with ID:", animal.ID)
	respondWithJSON(w, http.StatusOK, "ok", "Animal added successfully")
}

// respondWithJSON отправляет JSON-ответ
func respondWithJSON(w http.ResponseWriter, statusCode int, status, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  status,
		"message": message,
	})
}

// parseBool преобразует строковое значение в bool
func parseBool(value string) bool {
	result, _ := strconv.ParseBool(value)
	return result
}

// processRelation обрабатывает связанные данные: species, status, gender
func processRelation(db *gorm.DB, relation interface{}, column, value string) error {
	return db.Where(fmt.Sprintf("%s = ?", column), value).FirstOrCreate(relation).Error
}

// saveAnimalAge сохраняет возраст животного
func saveAnimalAge(db *gorm.DB, animal *models.Animal, yearsStr, monthsStr string) error {
	years, err := strconv.Atoi(yearsStr)
	if err != nil {
		years = 0
	}
	months, err := strconv.Atoi(monthsStr)
	if err != nil {
		months = 0
	}
	animalAge := models.AnimalAge{
		AnimalID: animal.ID,
		Years:    years,
		Months:   months,
	}
	return db.Save(&animalAge).Error
}

// validateAnimalImages проверяет изображения перед сохранением
func validateAnimalImages(r *http.Request) (string, error) {
	files := r.MultipartForm.File["images"]
	if len(files) > 4 {
		return "", fmt.Errorf("too many images uploaded")
	}
	for _, fileHeader := range files {
		fileExt := strings.ToLower(path.Ext(fileHeader.Filename))
		if !isValidImageExt(fileExt) {
			return fileExt, fmt.Errorf("invalid file type")
		}
	}
	return "", nil
}

func processAnimalImages(db *gorm.DB, r *http.Request, animalID uint) (string, error) {
	// Создаем директорию animals_images, если она не существует
	baseDir := "uploads/animal_images"
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
			log.Println("Error creating base directory:", err)
			return "", err
		}
	}

	animalDir := path.Join(baseDir, fmt.Sprintf("animal_%d", animalID))
	if _, err := os.Stat(animalDir); os.IsNotExist(err) {
		if err := os.MkdirAll(animalDir, os.ModePerm); err != nil {
			log.Println("Error creating animal directory:", err)
			return "", err
		}
	}

	files := r.MultipartForm.File["images"]
	log.Printf("Number of images: %d\n", len(files))
	if len(files) > 4 {
		log.Println("Too many images uploaded")
		return "", fmt.Errorf("You can upload a maximum of 4 images")
	}

	// Сохраняем каждое изображение
	for _, fileHeader := range files {
		if fileExt, err := saveImageAnimal(fileHeader, animalDir, animalID, db); err != nil {
			return fileExt, err
		}
	}
	return "", nil
}

// saveImageAnimal сохраняет одно изображение и создает запись в таблице PostImage
func saveImageAnimal(fileHeader *multipart.FileHeader, uploadDir string, animalID uint, db *gorm.DB) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		log.Println("Error opening file:", err)
		return "", err
	}
	defer file.Close()

	// Проверяем расширение файла
	fileExt := strings.ToLower(path.Ext(fileHeader.Filename))
	if !isValidImageExt(fileExt) {
		log.Println("Invalid file extension:", fileExt)
		return fileExt, fmt.Errorf("Invalid file type")
	}

	// Генерируем уникальное имя файла
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), fileExt)
	filePath := path.Join(uploadDir, fileName)

	// Создаем файл
	outFile, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating file:", err)
		return fileExt, err
	}
	defer outFile.Close()

	// Сохраняем содержимое
	if _, err := io.Copy(outFile, file); err != nil {
		log.Println("Error writing file:", err)
		return fileExt, err
	}

	// Сохраняем запись в базе данных
	image := models.PostImage{
		AnimalID: animalID,
		ImageURL: filepath.ToSlash(filePath),
	}
	if err := db.Create(&image).Error; err != nil {
		log.Println("Error saving image:", err)
		return fileExt, err
	}
	log.Println("Image saved successfully:", image.ImageURL)
	return fileExt, nil
}

// isValidImageExt проверяет допустимость расширения файла
func isValidImageExt(ext string) bool {
	validExts := []string{".jpg", ".jpeg", ".png", ".gif"}
	for _, v := range validExts {
		if ext == v {
			return true
		}
	}
	return false
}

// DeleteAnimal handles the deletion of an animal by ID
func DeleteAnimal(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	log.Println("Start DeleteAnimal handler")

	// Set response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Проверяем, что метод запроса — POST
	if r.Method != http.MethodPost {
		log.Println("Invalid request method")
		respondWithJSON(w, http.StatusMethodNotAllowed, "error", "Invalid request method")
		return
	}

	// Читаем тело запроса
	var requestData struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Println("Error decoding request body:", err)
		respondWithJSON(w, http.StatusBadRequest, "error", "Invalid request body")
		return
	}

	// Проверяем, что ID указан
	if requestData.ID == 0 {
		log.Println("Animal ID not provided")
		respondWithJSON(w, http.StatusBadRequest, "error", "Animal ID is required")
		return
	}

	animalID := requestData.ID

	// Fetch the animal to ensure it exists
	var animal models.Animal
	if err := db.Preload("Images").Preload("Age").Where("id = ?", animalID).First(&animal).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("Animal not found")
			respondWithJSON(w, http.StatusNotFound, "error", "Animal not found")
			return
		}
		log.Println("Error fetching animal:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error fetching animal")
		return
	}

	// Delete related images from the filesystem
	for _, image := range animal.Images {
		if err := os.Remove(image.ImageURL); err != nil && !os.IsNotExist(err) {
			log.Printf("Error deleting image file %s: %v\n", image.ImageURL, err)
		}
	}

	// Delete related records in the database
	if err := db.Where("animal_id = ?", animalID).Delete(&models.PostImage{}).Error; err != nil {
		log.Println("Error deleting related images from database:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error deleting related images")
		return
	}

	if err := db.Where("animal_id = ?", animalID).Delete(&models.AnimalAge{}).Error; err != nil {
		log.Println("Error deleting animal age:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error deleting animal age")
		return
	}

	// Delete the animal record
	if err := db.Delete(&animal).Error; err != nil {
		log.Println("Error deleting animal:", err)
		respondWithJSON(w, http.StatusInternalServerError, "error", "Error deleting animal")
		return
	}

	log.Printf("Animal with ID %d deleted successfully\n", animalID)
	respondWithJSON(w, http.StatusOK, "ok", "Animal deleted successfully")
}
