package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// AdminPanel displays the admin panel page with dynamic table data
func AdminPanel(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Get the current session
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	sessionToken := cookie.Value

	// Check if the session exists and retrieve user info
	var user User
	err = db.QueryRow("SELECT users.id, users.is_admin FROM users JOIN sessions ON users.id = sessions.user_id WHERE sessions.session_id = $1", sessionToken).Scan(&user.ID, &user.IsAdmin)
	if err != nil || !user.IsAdmin {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get the table parameter from the URL
	table := r.URL.Query().Get("table")
	fmt.Println("Table selected:", table) // Debugging statement

	if table == "" {
		http.Redirect(w, r, "/admin?table=users", http.StatusFound)
		return
	}

	var rows *sql.Rows

	// Dynamically fetch data from the selected table
	switch table {
	case "users":
		rows, err = db.Query("SELECT id, username, email, is_admin, role, email_confirmed FROM users")
	case "animals":
		rows, err = db.Query("SELECT id, name, age, breed, gender, arrival_date FROM animals")
	case "sessions":
		rows, err = db.Query("SELECT id, user_id, session_id, created_at, expires_at FROM sessions")
	default:
		http.Error(w, "Invalid table", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Prepare data for the template
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

	err = tmpl.Execute(w, map[string]interface{}{
		"Table": table,
		"Data":  data,
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
