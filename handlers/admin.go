package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
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
		rows, err = db.Query("SELECT id, username, email, is_admin FROM users")
	case "animals":
		rows, err = db.Query("SELECT id, name, species, age FROM animals")
	case "sessions":
		rows, err = db.Query("SELECT id, user_id, session_id FROM sessions")
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
