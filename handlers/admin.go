package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
)

// AdminPanel displays the admin panel page with dynamic table data
func AdminPanel(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем сессию пользователя
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	sessionToken := cookie.Value

	// Проверяем права пользователя
	var user User
	err = db.QueryRow("SELECT users.id, users.is_admin FROM users JOIN sessions ON users.id = sessions.user_id WHERE sessions.session_id = $1", sessionToken).Scan(&user.ID, &user.IsAdmin)
	if err != nil || !user.IsAdmin {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Получаем параметры таблицы и поиска
	table := r.URL.Query().Get("table")
	searchQuery := r.URL.Query().Get("search")

	if table == "" {
		http.Redirect(w, r, "/admin?table=users", http.StatusFound)
		return
	}

	var rows *sql.Rows
	query := ""

	// Формируем запрос на основе таблицы и поиска
	switch table {
	case "users":
		query = "SELECT id, username, email, is_admin, role, email_confirmed FROM users"
		if searchQuery != "" {
			query += " WHERE username ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%'"
			rows, err = db.Query(query, searchQuery)
		} else {
			rows, err = db.Query(query)
		}
	case "animals":
		query = "SELECT id, name, age, breed, gender, arrival_date FROM animals"
		if searchQuery != "" {
			query += " WHERE name ILIKE '%' || $1 || '%' OR breed ILIKE '%' || $1 || '%'"
			rows, err = db.Query(query, searchQuery)
		} else {
			rows, err = db.Query(query)
		}
	case "sessions":
		query = "SELECT id, user_id, session_id, created_at, expires_at FROM sessions"
		if searchQuery != "" {
			query += " WHERE session_id ILIKE '%' || $1 || '%'"
			rows, err = db.Query(query, searchQuery)
		} else {
			rows, err = db.Query(query)
		}
	default:
		http.Error(w, "Invalid table", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Обрабатываем данные для шаблона
	var data []map[string]interface{}
	columns, _ := rows.Columns()
	for rows.Next() {
		row := make(map[string]interface{})
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))

		for i := range values {
			pointers[i] = &values[i]
		}
		rows.Scan(pointers...)
		for i, col := range columns {
			row[col] = values[i]
		}
		data = append(data, row)
	}

	tmpl, err := template.ParseFiles("templates/admin.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Передаем данные в шаблон
	err = tmpl.Execute(w, map[string]interface{}{
		"Table":       table,
		"Data":        data,
		"SearchQuery": searchQuery,
	})
	if err != nil {
		http.Error(w, "Record not found", http.StatusInternalServerError)
	}
}

// DeleteRecord handles record deletion requests
func DeleteRecord(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the table and ID from the request
	table := r.URL.Query().Get("table")
	id := r.URL.Query().Get("id")

	if table == "" || id == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Convert ID to integer
	recordID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Delete record from the database
	var query string
	switch table {
	case "users":
		query = "DELETE FROM users WHERE id = $1"
	case "animals":
		query = "DELETE FROM animals WHERE id = $1"
	case "sessions":
		query = "DELETE FROM sessions WHERE id = $1"
	default:
		http.Error(w, "Invalid table", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(query, recordID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect back to the admin panel
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
