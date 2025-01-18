package handlers

import (
	"github.com/jinzhu/gorm"
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

func HomePage(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
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
func fetchAllAnimalsForList(db *gorm.DB) ([]AnimalForList, error) {
	var animals []AnimalForList

	err := db.Table("animals").
		Select("animals.id, animals.name, (SELECT image_url FROM postimages WHERE postimages.animal_id = animals.id LIMIT 1) AS image").
		Scan(&animals).Error

	if err != nil {
		return nil, err
	}

	return animals, nil
}
func TermsOfServicePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/user_agreement.html")
}
