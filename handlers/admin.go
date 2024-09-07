package handlers

import (
	"html/template"
	"net/http"
)

// AdminPanel отображает страницу панели администратора
func AdminPanel(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/admin_panel.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	// Отображаем панель администратора
	tmpl.Execute(w, nil)
}
