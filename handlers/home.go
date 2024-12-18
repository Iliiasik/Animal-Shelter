package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
)

type AnimalForList struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// PageData represents the data passed to the HTML templates
type PageData struct {
	LoggedIn bool
	Animals  []AnimalForList
}

// HomePage handles rendering the homepage
func HomePage(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	loggedIn := false

	// Проверяем сессию
	session, err := r.Cookie("session")
	if err == nil && session.Value != "" {
		loggedIn = true
	}

	// Получаем животных с деталями
	animals, err := fetchAllAnimalsForList(db)
	if err != nil {
		http.Error(w, "Error fetching animals", http.StatusInternalServerError)
		return
	}

	data := PageData{
		LoggedIn: loggedIn,
		Animals:  animals,
	}
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		return
	}
}

func fetchAllAnimalsForList(db *sql.DB) ([]AnimalForList, error) {
	var animals []AnimalForList

	// SQL-запрос для извлечения только необходимых данных
	query := `
		SELECT animals.id, animals.name, 
			(SELECT image_url FROM postimages WHERE postimages.animal_id = animals.id LIMIT 1) AS image
		FROM animals
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обрабатываем все строки результата
	for rows.Next() {
		var animal AnimalForList
		if err := rows.Scan(&animal.ID, &animal.Name, &animal.Image); err != nil {
			return nil, err
		}

		// Добавляем животное в список
		animals = append(animals, animal)
	}

	// Проверяем на наличие ошибок после завершения цикла
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return animals, nil
}

// TermsOfServicePage serves the terms of service page
func TermsOfServicePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/user_agreement.html")
}
