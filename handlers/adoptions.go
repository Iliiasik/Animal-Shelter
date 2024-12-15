package handlers

import (
	"database/sql"
	"encoding/json"
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

	// Проверяем, есть ли у этого пользователя заявка на это животное
	var userAdoptionExists int
	err = db.QueryRow(`SELECT COUNT(*) FROM adoptions WHERE animal_id = $1 AND user_id = $2 AND status_id IN (SELECT id FROM adoptionstatuses WHERE name IN ('Under review', 'Approved'))`, animalID, userID).Scan(&userAdoptionExists)
	if err != nil {
		sendAlert(w, Alert{"Error", "Error checking your adoption requests for this animal.", "error"})
		return
	}

	if userAdoptionExists > 0 {
		sendAlert(w, Alert{"Error", "You have already submitted an adoption request for this animal.", "warning"})
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
