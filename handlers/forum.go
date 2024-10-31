package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
)

var forumTemplates = template.Must(template.ParseFiles("templates/forum.html", "templates/new_topic.html", "templates/topic.html"))

func ShowForum(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем параметр поиска из запроса
	title := r.URL.Query().Get("title")

	// Основной SQL-запрос
	query := `
    SELECT t.id, t.title, u.username, u.profile_image, COUNT(p.id) AS response_count, 
           TO_CHAR(t.created_at, 'DD.MM.YYYY') AS created_at
    FROM topics t
    LEFT JOIN users u ON t.user_id = u.id
    LEFT JOIN posts p ON t.id = p.topic_id
    `

	// Если параметр title не пуст, добавляем условие WHERE
	var rows *sql.Rows
	var err error
	if title != "" {
		query += `WHERE t.title ILIKE $1 GROUP BY t.id, u.username, u.profile_image ORDER BY response_count DESC, t.created_at DESC`
		rows, err = db.Query(query, "%"+title+"%")
	} else {
		query += `GROUP BY t.id, u.username, u.profile_image ORDER BY response_count DESC, t.created_at DESC`
		rows, err = db.Query(query)
	}

	if err != nil {
		http.Error(w, "Error fetching topics", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Сканируем и отображаем результаты
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
			http.Error(w, "Error scanning topics", http.StatusInternalServerError)
			return
		}
		topics = append(topics, topic)
	}

	err = forumTemplates.ExecuteTemplate(w, "forum.html", topics)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
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
