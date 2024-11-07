package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

var forumTemplates = template.Must(template.ParseFiles("templates/forum.html", "templates/new_topic.html", "templates/topic.html"))

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

	// Получаем ID пользователя из cookie, если он авторизован
	cookie, err := r.Cookie("session")
	var userID int
	var userLoggedIn bool
	var userTopics, hottestTopics []struct {
		ID            int
		Title         string
		Username      string
		ProfileImage  string
		ResponseCount int
		CreatedAt     string
	}

	if err == nil {
		userLoggedIn = true
		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = $1", cookie.Value).Scan(&userID)
		if err != nil {
			log.Printf("Error fetching user ID: %v", err)
			return
		}

		// Запрос для тем пользователя
		userQuery := `
		SELECT t.id, t.title, u.username, u.profile_image, COUNT(p.id) AS response_count, 
		       TO_CHAR(t.created_at, 'DD.MM.YYYY') AS created_at
		FROM topics t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN posts p ON t.id = p.topic_id
		WHERE t.user_id = $1
		GROUP BY t.id, u.username, u.profile_image
		ORDER BY t.created_at DESC`

		rows, err := db.Query(userQuery, userID)
		if err != nil {
			log.Printf("Error fetching user topics: %v", err)
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
				log.Printf("Error scanning user topics: %v", err)
				return
			}
			userTopics = append(userTopics, topic)
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
	LIMIT 3`

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

	var topics []struct {
		ID            int
		Title         string
		Username      string
		ProfileImage  string
		ResponseCount int
		CreatedAt     string
	}

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
		UserTopics []struct {
			ID            int
			Title         string
			Username      string
			ProfileImage  string
			ResponseCount int
			CreatedAt     string
		}
		UserLoggedIn bool
		CurrentPage  int
		TotalPages   int
		Pages        []string
	}{
		Topics:        topics,
		HottestTopics: hottestTopics,
		UserTopics:    userTopics,
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
func compactPagination(current, total int, title string) []string {
	var pages []string

	// Функция для добавления номера страницы с учётом title
	addPageLink := func(page int) string {
		if title != "" {
			// Добавляем параметр title в URL
			return fmt.Sprintf("/forum?page=%d&title=%s", page, url.QueryEscape(title))
		}
		// Без параметра title
		return fmt.Sprintf("/forum?page=%d", page)
	}

	// Если всего страниц 5 или меньше, выводим все страницы
	if total <= 5 {
		for i := 1; i <= total; i++ {
			pages = append(pages, addPageLink(i))
		}
		return pages
	}

	// Добавляем первую страницу
	pages = append(pages, addPageLink(1))

	// Если текущая страница больше 3, добавляем троеточие
	if current > 3 {
		pages = append(pages, "...")
	}

	// Определяем диапазон для отображения номеров страниц вокруг текущей
	start := max(2, current-1)
	end := min(total-1, current+1)

	// Добавляем страницы в диапазоне от start до end
	for i := start; i <= end; i++ {
		pages = append(pages, addPageLink(i))
	}

	// Если текущая страница меньше чем total-2, добавляем троеточие
	if current < total-2 {
		pages = append(pages, "...")
	}

	// Добавляем последнюю страницу
	pages = append(pages, addPageLink(total))

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

// ShowNewTopicForm renders the form to create a new topic
func ShowNewTopicForm(w http.ResponseWriter, r *http.Request) {
	forumTemplates.ExecuteTemplate(w, "new_topic.html", nil)
}

// CreateTopic handles the creation of a new topic
func CreateTopic(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")

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

	// Проверяем, сколько топиков уже создал пользователь
	var topicCount int
	err = db.QueryRow("SELECT COUNT(*) FROM topics WHERE user_id = $1", userID).Scan(&topicCount)
	if err != nil {
		http.Error(w, "Error counting topics", http.StatusInternalServerError)
		return
	}

	if topicCount >= 3 { // Ограничение на 3 топика
		http.Error(w, "You can create a maximum of 3 topics", http.StatusForbidden)
		return
	}

	_, err = db.Exec("INSERT INTO topics (title, user_id) VALUES ($1, $2)", title, userID)
	if err != nil {
		http.Error(w, "Error creating topic", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/forum", http.StatusSeeOther)
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
