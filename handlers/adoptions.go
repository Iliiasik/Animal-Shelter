package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Alert struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Icon  string `json:"icon"`
}

func RegisterAdoption(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Проверяем, залогинен ли пользователь
	userID, userLoggedIn, err := getUserIDFromSession(db, r)
	if err != nil || !userLoggedIn {
		sendAlert(w, Alert{"Unauthorized", "Please log in to adopt an animal.", "error"})
		return
	}

	// Получаем animal_id из параметров запроса
	animalIDStr := r.URL.Query().Get("animal_id")
	if animalIDStr == "" {
		sendAlert(w, Alert{"Error", "Animal ID is required.", "error"})
		return
	}

	animalID, err := strconv.Atoi(animalIDStr)
	if err != nil {
		sendAlert(w, Alert{"Error", "Invalid Animal ID.", "error"})
		return
	}

	// Проверяем, является ли пользователь владельцем животного
	var ownerID int
	err = db.QueryRow("SELECT user_id FROM animals WHERE id = $1", animalID).Scan(&ownerID)
	if err != nil {
		sendAlert(w, Alert{"Error", "Animal not found.", "error"})
		return
	}

	if ownerID == userID {
		sendAlert(w, Alert{"Error", "You cannot adopt your own animal.", "error"})
		return
	}
	// Проверяем, есть ли у пользователя уже заявка на это животное
	var userAdoptionExists int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM adoptions 
		WHERE animal_id = $1 AND user_id = $2 AND status_id IN (SELECT id FROM adoptionstatuses WHERE name IN ('Under review', 'Approved'))
	`, animalID, userID).Scan(&userAdoptionExists)
	if err != nil {
		sendAlert(w, Alert{"Error", "Error checking your adoption requests for this animal.", "error"})
		return
	}

	if userAdoptionExists > 0 {
		sendAlert(w, Alert{"Error", "You have already submitted an adoption request for this animal.", "warning"})
		return
	}
	// Проверяем, есть ли уже активная заявка для этого животного
	var existingRequestCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM adoptions 
		WHERE animal_id = $1 AND status_id IN (SELECT id FROM adoptionstatuses WHERE name IN ('Under review', 'Approved'))
	`, animalID).Scan(&existingRequestCount)
	if err != nil {
		sendAlert(w, Alert{"Error", "Error checking existing adoption requests for this animal.", "error"})
		return
	}

	if existingRequestCount > 0 {
		sendAlert(w, Alert{"Error", "This animal already has an active adoption request.", "warning"})
		return
	}

	// Проверяем количество заявок пользователя
	var adoptionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM adoptions WHERE user_id = $1", userID).Scan(&adoptionCount)
	if err != nil {
		sendAlert(w, Alert{"Error", "Error checking your adoption requests.", "error"})
		return
	}

	if adoptionCount >= 5 {
		sendAlert(w, Alert{"Limit reached", "You have reached the maximum number of adoptions.", "warning"})
		return
	}

	// Получаем статус "Under review"
	var statusID uint
	err = db.QueryRow("SELECT id FROM adoptionstatuses WHERE name = $1", "Under review").Scan(&statusID)
	if err != nil {
		sendAlert(w, Alert{"Error", "Error fetching adoption status.", "error"})
		return
	}

	// Создаём заявку на усыновление
	_, err = db.Exec("INSERT INTO adoptions (animal_id, user_id, adoption_date, status_id) VALUES ($1, $2, $3, $4)", animalID, userID, time.Now(), statusID)
	if err != nil {
		sendAlert(w, Alert{"Error", "Error creating adoption request.", "error"})
		return
	}

	// Успешный ответ
	sendAlert(w, Alert{"Success", "Adoption request created successfully!", "success"})
}

// Функция для отправки JSON-ответа с уведомлением
func sendAlert(w http.ResponseWriter, alert Alert) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

func AcceptAdoption(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем adoption_id и animal_id из параметров запроса
	adoptionIDStr := r.URL.Query().Get("adoption_id")
	animalIDStr := r.URL.Query().Get("animal_id")
	if adoptionIDStr == "" || animalIDStr == "" {
		log.Printf("Error: Adoption ID and Animal ID are required. adoption_id: %s, animal_id: %s", adoptionIDStr, animalIDStr)
		sendAlert(w, Alert{"Error", "Adoption ID and Animal ID are required.", "error"})
		return
	}

	adoptionID, err := strconv.Atoi(adoptionIDStr)
	if err != nil {
		log.Printf("Error: Invalid Adoption ID. adoption_id: %s, error: %v", adoptionIDStr, err)
		sendAlert(w, Alert{"Error", "Invalid Adoption ID.", "error"})
		return
	}

	animalID, err := strconv.Atoi(animalIDStr)
	if err != nil {
		log.Printf("Error: Invalid Animal ID. animal_id: %s, error: %v", animalIDStr, err)
		sendAlert(w, Alert{"Error", "Invalid Animal ID.", "error"})
		return
	}

	// Проверяем, существует ли заявка с данным ID
	var adoptionStatusID int
	err = db.QueryRow("SELECT status_id FROM adoptions WHERE id = $1", adoptionID).Scan(&adoptionStatusID)
	if err != nil {
		log.Printf("Error: Adoption request not found. adoption_id: %d, error: %v", adoptionID, err)
		sendAlert(w, Alert{"Error", "Adoption request not found.", "error"})
		return
	}

	// Проверяем, что статус заявки "Under review"
	var underReviewStatusID int
	err = db.QueryRow("SELECT id FROM adoptionstatuses WHERE name = $1", "Under review").Scan(&underReviewStatusID)
	if err != nil {
		log.Printf("Error: Error fetching 'Under review' status. error: %v", err)
		sendAlert(w, Alert{"Error", "Error fetching 'Under review' status.", "error"})
		return
	}

	if adoptionStatusID != underReviewStatusID {
		log.Printf("Error: This adoption request is not under review. adoption_id: %d, status_id: %d", adoptionID, adoptionStatusID)
		sendAlert(w, Alert{"Error", "This adoption request is not under review.", "error"})
		return
	}

	// Обновляем статус заявки на "Approved"
	_, err = db.Exec("UPDATE adoptions SET status_id = (SELECT id FROM adoptionstatuses WHERE name = $1) WHERE id = $2", "Approved", adoptionID)
	if err != nil {
		log.Printf("Error: Error approving adoption request. adoption_id: %d, error: %v", adoptionID, err)
		sendAlert(w, Alert{"Error", "Error approving adoption request.", "error"})
		return
	}

	// Обновляем статус животного на "Booked"
	_, err = db.Exec("UPDATE animals SET status_id = (SELECT id FROM animalstatus WHERE name = $1) WHERE id = $2", "Booked", animalID)
	if err != nil {
		log.Printf("Error: Error updating animal status. animal_id: %d, error: %v", animalID, err)
		sendAlert(w, Alert{"Error", "Error updating animal status.", "error"})
		return
	}

	// Успешный ответ
	log.Printf("Success: Adoption request approved successfully. adoption_id: %d, animal_id: %d", adoptionID, animalID)
	sendAlert(w, Alert{"Success", "Adoption request approved successfully!", "success"})
}
func DeclineAdoption(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем adoption_id из параметров запроса
	adoptionIDStr := r.URL.Query().Get("adoption_id")
	if adoptionIDStr == "" {
		log.Printf("Error: Adoption ID is required. adoption_id: %s", adoptionIDStr)
		sendAlert(w, Alert{"Error", "Adoption ID is required.", "error"})
		return
	}

	adoptionID, err := strconv.Atoi(adoptionIDStr)
	if err != nil {
		log.Printf("Error: Invalid Adoption ID. adoption_id: %s, error: %v", adoptionIDStr, err)
		sendAlert(w, Alert{"Error", "Invalid Adoption ID.", "error"})
		return
	}

	// Проверяем, существует ли заявка с данным ID
	var adoptionStatusID int
	err = db.QueryRow("SELECT status_id FROM adoptions WHERE id = $1", adoptionID).Scan(&adoptionStatusID)
	if err != nil {
		log.Printf("Error: Adoption request not found. adoption_id: %d, error: %v", adoptionID, err)
		sendAlert(w, Alert{"Error", "Adoption request not found.", "error"})
		return
	}

	// Проверяем, что статус заявки "Under review"
	var underReviewStatusID int
	err = db.QueryRow("SELECT id FROM adoptionstatuses WHERE name = $1", "Under review").Scan(&underReviewStatusID)
	if err != nil {
		log.Printf("Error: Error fetching 'Under review' status. error: %v", err)
		sendAlert(w, Alert{"Error", "Error fetching 'Under review' status.", "error"})
		return
	}

	if adoptionStatusID != underReviewStatusID {
		log.Printf("Error: This adoption request is not under review. adoption_id: %d, status_id: %d", adoptionID, adoptionStatusID)
		sendAlert(w, Alert{"Error", "This adoption request is not under review.", "error"})
		return
	}

	// Удаляем заявку на усыновление
	_, err = db.Exec("DELETE FROM adoptions WHERE id = $1", adoptionID)
	if err != nil {
		log.Printf("Error: Error declining adoption request. adoption_id: %d, error: %v", adoptionID, err)
		sendAlert(w, Alert{"Error", "Error declining adoption request.", "error"})
		return
	}

	// Успешный ответ
	log.Printf("Success: Adoption request declined successfully. adoption_id: %d", adoptionID)
	sendAlert(w, Alert{"Success", "Adoption request declined successfully!", "success"})
}
