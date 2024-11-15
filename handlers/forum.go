package handlers

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

var forumTemplates = template.Must(template.ParseFiles("templates/forum.html", "templates/topic.html"))

func ShowForum(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из запроса
	title := r.URL.Query().Get("title")
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}

	const topicsPerPage = 16
	offset := (page - 1) * topicsPerPage
	type User struct {
		ID           int
		Username     string
		ProfileImage string
		BgImage      string
		Topics       []struct {
			ID            int
			Title         string
			ResponseCount int
			CreatedAt     string
		}
	}
	var (
		user                  User
		userLoggedIn          bool
		hottestTopics, topics []struct {
			ID            int
			Title         string
			Username      string
			ProfileImage  string
			ResponseCount int
			CreatedAt     string
		}
	)
	// Получаем ID пользователя из cookie, если он авторизован
	cookie, err := r.Cookie("session")
	if err == nil {
		userLoggedIn = true
		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = $1", cookie.Value).Scan(&user.ID)
		if err != nil {
			log.Printf("Error fetching user ID: %v", err)
			return
		}

		// Запрос для получения информации о пользователе и его темах
		err = db.QueryRow("SELECT username, profile_image, profile_bg_image FROM users WHERE id = $1", user.ID).Scan(&user.Username, &user.ProfileImage, &user.BgImage)
		if err != nil {
			log.Printf("Error fetching user info: %v", err)
			return
		}

		userQuery := `
            SELECT t.id, t.title, COUNT(p.id) AS response_count, TO_CHAR(t.created_at, 'DD.MM.YYYY') AS created_at
            FROM topics t
            LEFT JOIN posts p ON t.id = p.topic_id
            WHERE t.user_id = $1
            GROUP BY t.id
            ORDER BY t.id DESC
        `
		rows, err := db.Query(userQuery, user.ID)
		if err != nil {
			log.Printf("Error fetching user topics: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var topic struct {
				ID            int
				Title         string
				ResponseCount int
				CreatedAt     string
			}
			if err := rows.Scan(&topic.ID, &topic.Title, &topic.ResponseCount, &topic.CreatedAt); err != nil {
				log.Printf("Error scanning user topics: %v", err)
				return
			}
			user.Topics = append(user.Topics, topic)
		}
	}

	// Запрос для самых популярных тем
	hottestQuery := `
        SELECT t.id, t.title, u.username, u.profile_image, COUNT(p.id) AS response_count, 
               TO_CHAR(t.created_at, 'DD.MM.YYYY') AS created_at
        FROM topics t
        LEFT JOIN users u ON t.user_id = u.id
        LEFT JOIN posts p ON t.id = p.topic_id
        GROUP BY t.id, u.username, u.profile_image
        ORDER BY response_count DESC
        LIMIT 3
    `
	hottestRows, err := db.Query(hottestQuery)
	if err != nil {
		log.Printf("Error fetching hottest topics: %v", err)
		return
	}
	defer hottestRows.Close()

	for hottestRows.Next() {
		var topic struct {
			ID            int
			Title         string
			Username      string
			ProfileImage  string
			ResponseCount int
			CreatedAt     string
		}
		if err := hottestRows.Scan(&topic.ID, &topic.Title, &topic.Username, &topic.ProfileImage, &topic.ResponseCount, &topic.CreatedAt); err != nil {
			log.Printf("Error scanning hottest topics: %v", err)
			return
		}
		hottestTopics = append(hottestTopics, topic)
	}

	// Основной SQL-запрос для тем форума
	query := `
	SELECT t.id, t.title, u.username, u.profile_image, COUNT(p.id) AS response_count, 
	       TO_CHAR(t.created_at, 'DD.MM.YYYY') AS created_at
	FROM topics t
	LEFT JOIN users u ON t.user_id = u.id
	LEFT JOIN posts p ON t.id = p.topic_id`

	var rows *sql.Rows
	if title != "" {
		query += ` WHERE t.title ILIKE $1 GROUP BY t.id, u.username, u.profile_image ORDER BY t.id DESC, t.created_at DESC LIMIT $2 OFFSET $3`
		rows, err = db.Query(query, "%"+title+"%", topicsPerPage, offset)
	} else {
		query += ` GROUP BY t.id, u.username, u.profile_image ORDER BY t.id DESC, t.created_at DESC LIMIT $1 OFFSET $2`
		rows, err = db.Query(query, topicsPerPage, offset)
	}

	if err != nil {
		log.Printf("Error fetching topics: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var topic struct {
			ID            int
			Title         string
			Username      string
			ProfileImage  string
			ResponseCount int
			CreatedAt     string
		}
		if err := rows.Scan(&topic.ID, &topic.Title, &topic.Username, &topic.ProfileImage, &topic.ResponseCount, &topic.CreatedAt); err != nil {
			log.Printf("Error scanning topics: %v", err)
			return
		}
		topics = append(topics, topic)
	}

	// Получаем общее количество тем для пагинации
	var totalTopics int
	countQuery := `SELECT COUNT(*) FROM topics`
	if title != "" {
		countQuery += ` WHERE title ILIKE $1`
		err = db.QueryRow(countQuery, "%"+title+"%").Scan(&totalTopics)
	} else {
		err = db.QueryRow(countQuery).Scan(&totalTopics)
	}
	if err != nil {
		log.Printf("Error fetching total topic count: %v", err)
		return
	}

	pageCount := int(math.Ceil(float64(totalTopics) / float64(topicsPerPage)))

	// Формируем компактную пагинацию
	pages := compactPagination(page, pageCount, title)
	// Передаем данные в шаблон
	err = forumTemplates.ExecuteTemplate(w, "forum.html", struct {
		Topics []struct {
			ID            int
			Title         string
			Username      string
			ProfileImage  string
			ResponseCount int
			CreatedAt     string
		}
		HottestTopics []struct {
			ID            int
			Title         string
			Username      string
			ProfileImage  string
			ResponseCount int
			CreatedAt     string
		}
		User         User
		UserLoggedIn bool
		CurrentPage  int
		TotalPages   int
		Pages        []PageLink
	}{
		Topics:        topics,
		HottestTopics: hottestTopics,
		User:          user,
		UserLoggedIn:  userLoggedIn,
		CurrentPage:   page,
		TotalPages:    pageCount,
		Pages:         pages,
	})

	if err != nil {
		log.Printf("Error rendering template: %v", err)
	}
}

// compactPagination формирует список страниц в компактном виде
type PageLink struct {
	URL      string
	Number   string // текст для отображения на ссылке
	IsActive bool   // true для текущей страницы
}

func compactPagination(current, total int, title string) []PageLink {
	var pages []PageLink

	addPageLink := func(page int, isActive bool) PageLink {
		urlString := fmt.Sprintf("/forum?page=%d", page) // Используем urlString, чтобы избежать конфликта с пакетом
		if title != "" {
			// Используем правильный пакет для функции QueryEscape
			urlString += fmt.Sprintf("&title=%s", url.QueryEscape(title))
		}
		return PageLink{URL: urlString, Number: strconv.Itoa(page), IsActive: isActive}
	}

	// Если страниц меньше или равно 5, добавляем все страницы
	if total <= 5 {
		for i := 1; i <= total; i++ {
			pages = append(pages, addPageLink(i, i == current))
		}
		return pages
	}

	// Добавляем первую страницу
	pages = append(pages, addPageLink(1, current == 1))

	// Если текущая страница больше 3, добавляем троеточие
	if current > 3 {
		pages = append(pages, PageLink{Number: "..."})
	}

	// Определяем диапазон страниц вокруг текущей
	start := max(2, current-1)
	end := min(total-1, current+1)

	for i := start; i <= end; i++ {
		pages = append(pages, addPageLink(i, i == current))
	}

	// Если текущая страница меньше, чем total-2, добавляем троеточие
	if current < total-2 {
		pages = append(pages, PageLink{Number: "..."})
	}

	// Добавляем последнюю страницу
	pages = append(pages, addPageLink(total, current == total))

	return pages
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CreateTopic handles the creation of a new topic
func CreateTopic(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = $1", cookie.Value).Scan(&userID)
	if err != nil {
		http.Error(w, "Error fetching user ID", http.StatusInternalServerError)
		return
	}

	var topicCount int
	err = db.QueryRow("SELECT COUNT(*) FROM topics WHERE user_id = $1", userID).Scan(&topicCount)
	if err != nil {
		http.Error(w, "Error counting topics", http.StatusInternalServerError)
		return
	}

	if topicCount >= 10 {
		http.Error(w, "You can create a maximum of 10 topics", http.StatusForbidden)
		return
	}

	_, err = db.Exec("INSERT INTO topics (title, description, user_id) VALUES ($1, $2, $3)", title, description, userID)
	if err != nil {
		http.Error(w, "Error creating topic", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // Возвращает 200 при успешном создании
}

// ShowTopic displays the messages in a topic
func ShowTopic(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	topicID := r.URL.Query().Get("id")

	// Получаем заголовок темы
	var title string
	err := db.QueryRow("SELECT title FROM topics WHERE id = $1", topicID).Scan(&title)
	if err != nil {
		http.Error(w, "Error fetching topic title", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, content, user_id, created_at FROM posts WHERE topic_id = $1 ORDER BY created_at", topicID)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []struct {
		ID        int
		Content   string
		UserID    int
		CreatedAt string
	}

	for rows.Next() {
		var post struct {
			ID        int
			Content   string
			UserID    int
			CreatedAt string
		}
		if err := rows.Scan(&post.ID, &post.Content, &post.UserID, &post.CreatedAt); err != nil {
			http.Error(w, "Error scanning posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	// Преобразование topicID в int
	topicIDInt, err := strconv.Atoi(topicID)
	if err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	// Передаем заголовок и посты в шаблон
	err = forumTemplates.ExecuteTemplate(w, "topic.html", struct {
		Title string
		ID    int
		Posts []struct {
			ID        int
			Content   string
			UserID    int
			CreatedAt string
		}
	}{
		Title: title,
		ID:    topicIDInt, // Передаем ID темы как int
		Posts: posts,
	})

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// CreatePost handles adding a new post to a topic
func CreatePost(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topicID := r.FormValue("topic_id")
	content := r.FormValue("content")

	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = $1", cookie.Value).Scan(&userID)
	if err != nil {
		http.Error(w, "Error fetching user ID", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO posts (topic_id, user_id, content) VALUES ($1, $2, $3)", topicID, userID, content)
	if err != nil {
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/topic?id="+topicID, http.StatusSeeOther)
}

func DeleteTopics(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	// Проверка метода запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим форму
	if err := r.ParseForm(); err != nil {
		log.Println("Error parsing form:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Получаем список ID топиков из формы
	topicIDs := r.Form["topic_ids[]"] // Обрати внимание на использование topic_ids[]

	// Выводим полученные значения для отладки
	//log.Println("Received topic IDs: ", topicIDs)

	// Если не выбраны никакие топики
	if len(topicIDs) == 0 {
		http.Error(w, "Please select at least one topic to delete", http.StatusBadRequest)
		return
	}

	// Начинаем транзакцию
	tx := db.Begin()
	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Выполняем каскадное удаление для каждого ID
	for _, id := range topicIDs {
		// Удаляем связанные записи из таблицы posts
		if err := tx.Exec("DELETE FROM posts WHERE topic_id = ?", id).Error; err != nil {
			tx.Rollback()
			log.Println("Error deleting related records:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Удаляем сам топик
		if err := tx.Exec("DELETE FROM topics WHERE id = ?", id).Error; err != nil {
			tx.Rollback()
			log.Println("Error deleting topic:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Подтверждаем транзакцию
	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Перенаправляем пользователя после успешного удаления
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}
