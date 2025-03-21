package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
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
	userID, userLoggedIn, err := getUserIDFromSession(db, r)
	if err != nil {
		log.Printf("Error fetching user ID from session: %v", err)
		return
	}

	if userLoggedIn {
		// Получаем информацию о пользователе
		err := db.QueryRow(`SELECT username FROM users WHERE id = $1`, userID).Scan(&user.Username)
		if err != nil {
			log.Printf("Error fetching user info: %v", err)
			return
		}

		// Запрос для получения изображений профиля и фона
		err = db.QueryRow(`SELECT profile_image, profile_bg_image FROM user_images WHERE user_id = $1`, userID).Scan(&user.ProfileImage, &user.BgImage)
		if err != nil {
			log.Printf("Error fetching user images: %v", err)
			return
		}

		// Запрос для получения тем пользователя
		userQuery := `
			SELECT t.id, t.title, COUNT(p.id) AS response_count, TO_CHAR(t.created_at, 'DD.MM.YYYY') AS created_at
			FROM topics t
			LEFT JOIN posts p ON t.id = p.topic_id
			WHERE t.user_id = $1
			GROUP BY t.id
			ORDER BY t.id DESC
		`
		stmt, err := db.Prepare(userQuery)
		if err != nil {
			log.Printf("Error preparing user query: %v", err)
			return
		}
		defer stmt.Close()

		rows, err := stmt.Query(userID)
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
        SELECT t.id, t.title, u.username, ui.profile_image, COUNT(p.id) AS response_count, 
            TO_CHAR(t.created_at, 'DD.MM.YYYY') AS created_at
        FROM topics t
        LEFT JOIN users u ON t.user_id = u.id
        LEFT JOIN user_images ui ON u.id = ui.user_id
        LEFT JOIN posts p ON t.id = p.topic_id
        GROUP BY t.id, u.username, ui.profile_image
        ORDER BY response_count DESC
        LIMIT 3
    `
	hottestStmt, err := db.Prepare(hottestQuery)
	if err != nil {
		log.Printf("Error preparing hottest query: %v", err)
		return
	}
	defer hottestStmt.Close()

	hottestRows, err := hottestStmt.Query()
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
        SELECT t.id, t.title, u.username, ui.profile_image, COUNT(p.id) AS response_count, 
            TO_CHAR(t.created_at, 'DD.MM.YYYY') AS created_at
        FROM topics t
        LEFT JOIN users u ON t.user_id = u.id
        LEFT JOIN user_images ui ON u.id = ui.user_id
        LEFT JOIN posts p ON t.id = p.topic_id
    `

	var rows *sql.Rows
	if title != "" {
		query += ` WHERE t.title ILIKE $1 GROUP BY t.id, u.username, ui.profile_image ORDER BY t.id DESC, t.created_at DESC LIMIT $2 OFFSET $3`
		stmt, err := db.Prepare(query)
		if err != nil {
			log.Printf("Error preparing query for topics: %v", err)
			return
		}
		defer stmt.Close()

		rows, err = stmt.Query("%"+title+"%", topicsPerPage, offset)
	} else {
		query += ` GROUP BY t.id, u.username, ui.profile_image ORDER BY t.id DESC, t.created_at DESC LIMIT $1 OFFSET $2`
		stmt, err := db.Prepare(query)
		if err != nil {
			log.Printf("Error preparing query for topics: %v", err)
			return
		}
		defer stmt.Close()

		rows, err = stmt.Query(topicsPerPage, offset)
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
		stmt, err := db.Prepare(countQuery)
		if err != nil {
			log.Printf("Error preparing count query: %v", err)
			return
		}
		defer stmt.Close()

		err = stmt.QueryRow("%" + title + "%").Scan(&totalTopics)
	} else {
		stmt, err := db.Prepare(countQuery)
		if err != nil {
			log.Printf("Error preparing count query: %v", err)
			return
		}
		defer stmt.Close()

		err = stmt.QueryRow().Scan(&totalTopics)
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
		log.Printf("Error executing template: %v", err)
	}
}

type PageLink struct {
	URL      string
	Number   string
	IsActive bool
}

func compactPagination(current, total int, title string) []PageLink {
	var pages []PageLink

	addPageLink := func(page int, isActive bool) PageLink {
		urlString := fmt.Sprintf("/forum?page=%d", page)
		if title != "" {
			urlString += fmt.Sprintf("&title=%s", url.QueryEscape(title))
		}
		return PageLink{URL: urlString, Number: strconv.Itoa(page), IsActive: isActive}
	}

	if total <= 5 {
		for i := 1; i <= total; i++ {
			pages = append(pages, addPageLink(i, i == current))
		}
		return pages
	}

	pages = append(pages, addPageLink(1, current == 1))

	if current > 3 {
		pages = append(pages, PageLink{Number: "..."})
	}

	start := max(2, current-1)
	end := min(total-1, current+1)

	for i := start; i <= end; i++ {
		pages = append(pages, addPageLink(i, i == current))
	}

	if current < total-2 {
		pages = append(pages, PageLink{Number: "..."})
	}

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
func CreateTopic(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	// Получаем userID из сессии с помощью функции getUserIDFromSession
	userID, loggedIn, err := getUserIDFromSession(db, r)
	if err != nil {
		http.Error(w, "Error fetching user ID", http.StatusInternalServerError)
		return
	}
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Подготавливаем запрос для подсчета количества тем
	stmtCount, err := db.Prepare("SELECT COUNT(*) FROM topics WHERE user_id = $1")
	if err != nil {
		http.Error(w, "Error preparing statement for topic count", http.StatusInternalServerError)
		return
	}
	defer stmtCount.Close()

	var topicCount int
	err = stmtCount.QueryRow(userID).Scan(&topicCount)
	if err != nil {
		http.Error(w, "Error counting topics", http.StatusInternalServerError)
		return
	}

	if topicCount >= 10 {
		http.Error(w, "You can create a maximum of 10 topics", http.StatusForbidden)
		return
	}

	// Подготавливаем запрос для вставки новой темы
	stmtInsert, err := db.Prepare("INSERT INTO topics (title, description, user_id) VALUES ($1, $2, $3)")
	if err != nil {
		http.Error(w, "Error preparing statement for topic insertion", http.StatusInternalServerError)
		return
	}
	defer stmtInsert.Close()

	// Выполняем вставку
	_, err = stmtInsert.Exec(title, description, userID)
	if err != nil {
		http.Error(w, "Error creating topic", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ об успешном создании
	w.WriteHeader(http.StatusOK) // Возвращает 200 при успешном создании
}

func ShowTopic(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	topicID := r.URL.Query().Get("id")

	// Преобразуем topicID в int
	topicIDInt, err := strconv.Atoi(topicID)
	if err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	// Подготавливаем запрос для получения заголовка, описания и информации о создателе темы
	stmtTopic, err := db.Prepare(
		`SELECT t.title, t.description, u.username, ui.profile_image 
		FROM topics t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN user_images ui ON u.id = ui.user_id
		WHERE t.id = $1`)
	if err != nil {
		http.Error(w, "Error preparing statement for topic", http.StatusInternalServerError)
		return
	}
	defer stmtTopic.Close()

	var title, description, creatorUsername, creatorProfileImage string
	err = stmtTopic.QueryRow(topicIDInt).Scan(&title, &description, &creatorUsername, &creatorProfileImage)
	if err != nil {
		http.Error(w, "Error fetching topic details", http.StatusInternalServerError)
		return
	}

	// Подготовка запроса для получения постов по ID темы
	stmtPosts, err := db.Prepare(
		`SELECT id, content, user_id, created_at, parent_id, rating, image_url 
		FROM posts 
		WHERE topic_id = $1 
		ORDER BY created_at`)
	if err != nil {
		http.Error(w, "Error preparing statement for posts", http.StatusInternalServerError)
		return
	}
	defer stmtPosts.Close()

	rows, err := stmtPosts.Query(topicIDInt)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Структура для постов с дополнительной информацией
	type Post struct {
		ID           int
		Content      string
		UserID       int
		CreatedAt    string
		ParentID     *int
		Username     string
		ProfileImage string
		Replies      []Post
		TopicID      int // Добавляем поле TopicID
		Rating       int
		ImageURL     string
	}

	var posts []Post
	var postMap = make(map[int][]Post) // Карта для хранения ответов на посты

	// Получаем все посты и ответы
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Content, &post.UserID, &post.CreatedAt, &post.ParentID, &post.Rating, &post.ImageURL); err != nil {
			http.Error(w, "Error scanning posts", http.StatusInternalServerError)
			return
		}

		// Устанавливаем TopicID для каждого поста
		post.TopicID = topicIDInt

		// Получение username и ProfileImage для каждого поста
		err = db.QueryRow(
			`SELECT u.username, ui.profile_image 
			FROM users u 
			LEFT JOIN user_images ui ON u.id = ui.user_id 
			WHERE u.id = $1`,
			post.UserID,
		).Scan(&post.Username, &post.ProfileImage)

		if err != nil {
			http.Error(w, "Error fetching user details", http.StatusInternalServerError)
			return
		}

		// Добавляем пост в основной список
		if post.ParentID == nil {
			posts = append(posts, post)
		} else {
			// Сохраняем ответы на другие посты
			postMap[*post.ParentID] = append(postMap[*post.ParentID], post)
		}
	}

	// Привязываем ответы к родительским постам
	for i := range posts {
		posts[i].Replies = postMap[posts[i].ID]
	}

	// Получаем userID из сессии
	userID, userLoggedIn, err := getUserIDFromSession(db, r)
	if err != nil {
		http.Error(w, "Error getting user ID from session", http.StatusInternalServerError)
		return
	}

	// Проверяем, есть ли лайк на тему
	var isLiked bool
	if userLoggedIn {
		// Проверяем, поставил ли пользователь лайк
		err = db.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND topic_id = $2)`,
			userID, topicIDInt,
		).Scan(&isLiked)
		if err != nil {
			http.Error(w, "Error checking like status", http.StatusInternalServerError)
			return
		}
	}
	var likeCount int
	err = db.QueryRow(
		`SELECT COUNT(*) FROM likes WHERE topic_id = $1`,
		topicIDInt,
	).Scan(&likeCount)
	if err != nil {
		http.Error(w, "Error fetching like count", http.StatusInternalServerError)
		return
	}

	// Запрос для поиска похожих тем (без ResponseCount и CreatedAt)
	var similarTopics []struct {
		ID           int
		Title        string
		Username     string
		ProfileImage string
	}
	similarTopicsQuery := `
		SELECT t.id, t.title, u.username, ui.profile_image
		FROM topics t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN user_images ui ON u.id = ui.user_id
		WHERE t.title ILIKE $1 AND t.id != $2
		ORDER BY t.created_at DESC
		LIMIT 5
	`
	rows, err = db.Query(similarTopicsQuery, "%"+title+"%", topicIDInt)
	if err != nil {
		http.Error(w, "Error fetching similar topics", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var similarTopic struct {
			ID           int
			Title        string
			Username     string
			ProfileImage string
		}
		if err := rows.Scan(&similarTopic.ID, &similarTopic.Title, &similarTopic.Username, &similarTopic.ProfileImage); err != nil {
			http.Error(w, "Error scanning similar topics", http.StatusInternalServerError)
			return
		}
		similarTopics = append(similarTopics, similarTopic)
	}

	// Передаем данные в шаблон с IDt для темы
	err = forumTemplates.ExecuteTemplate(w, "topic.html", struct {
		Title               string
		Description         string
		CreatorUsername     string
		CreatorProfileImage string
		IDt                 int
		Posts               []Post
		IsLiked             bool
		LikeCount           int
		SimilarTopics       []struct {
			ID           int
			Title        string
			Username     string
			ProfileImage string
		}
	}{
		Title:               title,
		Description:         description,
		CreatorUsername:     creatorUsername,
		CreatorProfileImage: creatorProfileImage,
		IDt:                 topicIDInt, // Передаем topic ID как IDt
		Posts:               posts,
		IsLiked:             isLiked,
		LikeCount:           likeCount,
		SimilarTopics:       similarTopics,
	})

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func CreatePost(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart/form-data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Unable to process form", http.StatusBadRequest)
		return
	}

	// Получаем данные из формы
	topicID := r.FormValue("topic_id")
	content := r.FormValue("content")
	parentID := r.FormValue("comment_id")

	if topicID == "" || content == "" {
		http.Error(w, "Topic ID and content are required", http.StatusBadRequest)
		return
	}

	var parentIDInt *int
	if parentID != "" {
		parsedParentID, err := strconv.Atoi(parentID)
		if err != nil {
			log.Printf("Invalid ParentID %s: %v", parentID, err)
			http.Error(w, "Parent comment ID must be a number", http.StatusBadRequest)
			return
		}
		parentIDInt = &parsedParentID
	}

	// Получаем userID из сессии
	userID, isAuthorized, err := getUserIDFromSession(db, r)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Обработка загруженного файла
	var photoPath string
	file, handler, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()
		photoPath, err = SaveUploadedFile(file, handler, topicID, userID)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
	} else {
		log.Printf("No file uploaded or error: %v", err)
		photoPath = ""
	}

	// Вставка поста в базу данных
	stmtInsertPost, err := db.Prepare(`
        INSERT INTO posts (topic_id, user_id, content, parent_id, image_url) 
        VALUES ($1, $2, $3, $4, $5) RETURNING id
    `)
	if err != nil {
		log.Printf("Error preparing statement for inserting post: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer stmtInsertPost.Close()

	var postID int
	err = stmtInsertPost.QueryRow(topicID, userID, content, parentIDInt, photoPath).Scan(&postID)
	if err != nil {
		log.Printf("Error during post creation: %v", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	log.Printf("New post created with ID: %d", postID)
	http.Redirect(w, r, fmt.Sprintf("/topic?id=%s", topicID), http.StatusSeeOther)
}
func SaveUploadedFile(file multipart.File, handler *multipart.FileHeader, topicID string, userID int) (string, error) {
	// Создаём путь к директории для топика с приставкой "topic_"
	topicDir := fmt.Sprintf("topic_%s", topicID)
	uploadDir := path.Join("uploads/forum_images", topicDir, fmt.Sprintf("user_%d", userID)) // Создаём директорию для пользователя внутри топика

	// Создаём директорию для пользователя в топике, если её ещё нет
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			log.Printf("Error creating user directory in topic: %v", err)
			return "", err
		}
	}

	// Генерация уникального имени файла
	fileUUID := uuid.New().String()
	fileExtension := filepath.Ext(handler.Filename)          // Расширение файла
	fileName := fmt.Sprintf("%s%s", fileUUID, fileExtension) // Используем только UUID для имени файла

	// Формирование полного пути для сохранения файла
	photoPath := path.Join(uploadDir, fileName)

	// Сохраняем файл
	destFile, err := os.Create(photoPath)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		return "", err
	}
	defer destFile.Close()

	// Копируем содержимое файла в созданный файл
	if _, err := io.Copy(destFile, file); err != nil {
		log.Printf("Error copying file: %v", err)
		return "", err
	}

	log.Printf("File uploaded successfully: %s", photoPath)
	return photoPath, nil
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
		// Удаляем связанные лайки
		if err := tx.Exec("DELETE FROM likes WHERE topic_id = ?", id).Error; err != nil {
			tx.Rollback()
			log.Println("Error deleting likes:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

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
func getUserIDFromSession(db *sql.DB, r *http.Request) (int, bool, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		// Если куки нет, значит пользователь не авторизован
		return 0, false, nil
	}

	// Если кука существует, проверяем сессию
	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = $1", cookie.Value).Scan(&userID)
	if err != nil {
		log.Printf("Error fetching user ID: %v", err)
		return 0, false, err
	}

	return userID, true, nil
}
func ToggleLike(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор пользователя из сессии
	userID, isAuthorized, err := getUserIDFromSession(db, r)
	if err != nil {
		http.Error(w, "Error fetching user ID", http.StatusInternalServerError)
		return
	}

	// Если пользователь не авторизован, перенаправляем на страницу входа
	if !isAuthorized {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем topic_id из URL
	topicID := r.URL.Query().Get("topic_id")
	if topicID == "" {
		http.Error(w, "Topic ID is required", http.StatusBadRequest)
		return
	}
	topicIDInt, err := strconv.Atoi(topicID)
	if err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	// Проверка, существует ли лайк
	var likeID int
	err = db.QueryRow(
		"SELECT id FROM likes WHERE user_id = $1 AND topic_id = $2",
		userID, topicIDInt,
	).Scan(&likeID)

	if err == nil { // Лайк существует, удаляем
		_, err := db.Exec("DELETE FROM likes WHERE id = $1", likeID)
		if err != nil {
			http.Error(w, "Error deleting like", http.StatusInternalServerError)
			log.Printf("Error deleting like: %v", err) // Логирование ошибки удаления
			return
		}
		log.Printf("Like deleted for topic_id: %d by user_id: %d", topicIDInt, userID)
	} else if err == sql.ErrNoRows { // Лайк не существует, добавляем
		_, err := db.Exec(
			"INSERT INTO likes (user_id, topic_id) VALUES ($1, $2)",
			userID, topicIDInt,
		)
		if err != nil {
			http.Error(w, "Error adding like", http.StatusInternalServerError)
			log.Printf("Error adding like: %v", err) // Логирование ошибки вставки
			return
		}
		log.Printf("Like added for topic_id: %d by user_id: %d", topicIDInt, userID)
	} else {
		http.Error(w, "Error checking like existence", http.StatusInternalServerError)
		log.Printf("Error checking like existence: %v", err) // Логирование ошибки запроса
		return
	}

	// Перенаправляем на страницу темы, чтобы обновить состояние
	http.Redirect(w, r, fmt.Sprintf("/topic?id=%d", topicIDInt), http.StatusSeeOther)
}
func UpdateRating(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из запроса
	postID := r.URL.Query().Get("post_id")
	action := r.URL.Query().Get("action") // "like" или "dislike"

	// Получаем user_id из сессии
	userID, loggedIn, err := getUserIDFromSession(db, r)
	if err != nil {
		http.Error(w, "Error retrieving user ID from session", http.StatusInternalServerError)
		return
	}
	if !loggedIn {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	// Преобразуем postID в int
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Проверяем, есть ли уже лайк или дизлайк от этого пользователя для данного поста
	var existingLikeStatus sql.NullBool
	err = db.QueryRow("SELECT like_status FROM post_likes WHERE post_id = $1 AND user_id = $2", postIDInt, userID).Scan(&existingLikeStatus)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Error checking existing like status", http.StatusInternalServerError)
		return
	}

	// Логика для обработки лайков и дизлайков
	if existingLikeStatus.Valid {
		// Если уже есть лайк/дизлайк, меняем его
		if action == "like" && !existingLikeStatus.Bool {
			_, err = db.Exec("UPDATE post_likes SET like_status = true WHERE post_id = $1 AND user_id = $2", postIDInt, userID)
			if err != nil {
				http.Error(w, "Error updating like status", http.StatusInternalServerError)
				return
			}
		} else if action == "dislike" && existingLikeStatus.Bool {
			_, err = db.Exec("UPDATE post_likes SET like_status = false WHERE post_id = $1 AND user_id = $2", postIDInt, userID)
			if err != nil {
				http.Error(w, "Error updating dislike status", http.StatusInternalServerError)
				return
			}
		} else {
			// Если выбран тот же статус, то ничего не меняем
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "No change to the rating")
			return
		}
	} else {
		// Если лайк/дизлайк еще не был поставлен, добавляем его
		if action == "like" {
			_, err = db.Exec("INSERT INTO post_likes (post_id, user_id, like_status) VALUES ($1, $2, true)", postIDInt, userID)
			if err != nil {
				http.Error(w, "Error inserting like status", http.StatusInternalServerError)
				return
			}
		} else if action == "dislike" {
			_, err = db.Exec("INSERT INTO post_likes (post_id, user_id, like_status) VALUES ($1, $2, false)", postIDInt, userID)
			if err != nil {
				http.Error(w, "Error inserting dislike status", http.StatusInternalServerError)
				return
			}
		}
	}

	// Пересчитываем общий рейтинг поста (обновляем его в таблице)
	var totalLikes, totalDislikes int
	err = db.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = $1 AND like_status = true", postIDInt).Scan(&totalLikes)
	if err != nil {
		http.Error(w, "Error fetching total likes", http.StatusInternalServerError)
		return
	}
	err = db.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = $1 AND like_status = false", postIDInt).Scan(&totalDislikes)
	if err != nil {
		http.Error(w, "Error fetching total dislikes", http.StatusInternalServerError)
		return
	}

	// Обновляем рейтинг поста
	finalRating := totalLikes - totalDislikes
	_, err = db.Exec("UPDATE posts SET rating = $1 WHERE id = $2", finalRating, postIDInt)
	if err != nil {
		http.Error(w, "Error updating post rating", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Rating updated to %d", finalRating)
}
