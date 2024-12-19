package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("mainpage.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Обработчик на страницу регистрации signup GET
func SignUpGetHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("signup.html") // путь к странице регистрации
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Обработчик для регистрации пользователя signup POST
func SignUpPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req SignUpRequest
		// Логирование входящего запроса
		log.Println("Received request for signup")

		// Декодируем тело запроса
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		log.Printf("Decoded data: %+v", req)

		// Создание нового пользователя
		user := User{
			Email:    req.Email,
			Username: req.Username,
			Password: req.Password, // Убедитесь, что пароль хэшируется
		}
		log.Printf("Created user object: %+v", user)

		// Вставка пользователя в базу данных
		err := insertUser(user)
		if err != nil {
			log.Printf("Error inserting user: %v", err)

			if errors.Is(err, ErrDuplicateKey) {
				// Отправляем статус 409 в случае дублирования
				http.Error(w, "Account with this username or email already exists", http.StatusConflict)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Успех
		log.Println("User successfully registered")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User registered successfully"))
	}
}

// Обработчик на страницу регистрации login GET
func LogInGetHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("login.html") // путь к странице регистрации
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	// Логирование входящего запроса
	log.Println("Received request for login")

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user := User{
		Username: req.Username,
		Password: req.Password,
	}
	log.Printf("Decoded data: %+v", req)
	exists, userID, err := findUser(&user)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}
	if !exists {
		handleError(w, fmt.Errorf("user not found"), http.StatusNotFound)
		return
	}

	accessToken, err := generateAccessToken(strconv.Itoa(userID))
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	refreshToken, err := generateRefreshToken(strconv.Itoa(userID))
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: false,
	})
	log.Printf("Set refresh token: %s", refreshToken)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	fmt.Fprintf(w, `{"accessToken": "%s"}`, accessToken)
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение refresh-токена из куков
	refreshCookie, err := r.Cookie("refreshToken")
	if err != nil {
		http.Error(w, "Refresh token is missing", http.StatusUnauthorized)
		log.Printf("ошибка 1 ")
		return
	}
	refreshToken := refreshCookie.Value

	// 2. Проверка refresh-токена
	claims, err := validateToken(refreshToken, refreshSecret)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// 3. Получение userID из claims
	userID, ok := claims["userID"].(string)
	if !ok {
		http.Error(w, "Invalid claims in refresh token", http.StatusUnauthorized)
		return
	}

	// 4. Генерация нового access-токена
	newAccessToken, err := generateAccessToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	// 5. Генерация нового refresh-токена
	newRefreshToken, err := generateRefreshToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// 6. Обновление refresh-токена в куках
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    newRefreshToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   false, // Убедись, что используешь HTTPS
		SameSite: http.SameSiteNoneMode,
	})

	// 7. Отправка нового access-токена клиенту
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	response := struct {
		AccessToken string `json:"accessToken"`
	}{
		AccessToken: newAccessToken,
	}

	json.NewEncoder(w).Encode(response)
}

// Handler для получения блокнотов пользователя
func getNotebooksHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем токен из заголовка Authorization
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Убираем "Bearer " из начала строки, если он есть
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[len("Bearer "):]
	}

	// Проверяем и валидируем токен
	claims, err := validateToken(tokenString, accessSecret)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Извлекаем userID из токена
	userIDString, ok := claims["userID"].(string)
	if !ok {
		http.Error(w, "Invalid userID in token", http.StatusUnauthorized)
		return
	}

	// Преобразуем userID из строки в int
	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		http.Error(w, "Invalid userID format", http.StatusUnauthorized)
		return
	}

	// Получаем блокноты пользователя
	notebooks, err := getNotebooksByUserID(userID)
	if err != nil {
		log.Println("Error fetching notebooks:", err) // Логируем ошибку
		http.Error(w, "Failed to fetch notebooks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notebooks)
}

// createNotebookHandler — обработчик для создания нового блокнота
func createNotebookHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем токен из заголовка Authorization
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Убираем "Bearer " из начала строки, если он есть
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[len("Bearer "):]
	}

	// Проверяем и валидируем токен
	claims, err := validateToken(tokenString, accessSecret)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Извлекаем userID из токена
	userIDString, ok := claims["userID"].(string)
	if !ok {
		http.Error(w, "Invalid userID in token", http.StatusUnauthorized)
		return
	}

	// Преобразуем userID из строки в int
	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		http.Error(w, "Invalid userID format", http.StatusUnauthorized)
		return
	}

	// Парсим данные о блокноте из тела запроса
	var notebook struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&notebook); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Проверяем имя блокнота на пустое значение
	if notebook.Name == "" {
		http.Error(w, "Notebook name is required", http.StatusBadRequest)
		return
	}

	// Создаем структуру для вставки в базу данных
	notebookToInsert := Notebook{
		UserID:    userID, // Используем преобразованный userID как int
		Name:      notebook.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Вставляем блокнот в базу данных
	err = insertNotebook(notebookToInsert)
	if err != nil {
		http.Error(w, "Failed to create notebook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
		ID      int    `json:"id"`
		Name    string `json:"name"`
	}{
		Message: "Notebook created successfully",
		ID:      notebookToInsert.ID, // ID блока после вставки
		Name:    notebookToInsert.Name,
	})
}

// Обработчик для обновления блокнота
func updateNotebookHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор блокнота из URL
	notebookIDStr := r.URL.Path[len("/api/notebooks/"):]

	if notebookIDStr == "" {
		http.Error(w, "Missing notebook_id in URL path", http.StatusBadRequest)
		return
	}

	notebookID, err := strconv.Atoi(notebookIDStr)
	if err != nil {
		http.Error(w, "Invalid notebook_id format", http.StatusBadRequest)
		return
	}

	// Читаем данные из тела запроса
	notebook := Notebook{}
	err = json.NewDecoder(r.Body).Decode(&notebook)
	notebook.ID = notebookID
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Выполняем обновление в базе данных
	err = UpdateNotebook(notebook)
	if err != nil {
		http.Error(w, "Failed to update notebook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notebook updated"})
}

func deleteNotebookHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор блокнота из URL
	notebookIDStr := r.URL.Path[len("/api/notebooks/"):]

	if notebookIDStr == "" {
		http.Error(w, "Missing notebook_id in URL path", http.StatusBadRequest)
		return
	}

	notebookID, err := strconv.Atoi(notebookIDStr)
	if err != nil {
		http.Error(w, "Invalid notebook_id format", http.StatusBadRequest)
		return
	}

	// Читаем данные из тела запроса
	notebook := Notebook{}
	//err = json.NewDecoder(r.Body).Decode(&notebook)
	notebook.ID = notebookID
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Выполняем обновление в базе данных
	err = DeleteNotebook(notebook)
	if err != nil {
		http.Error(w, "Failed to update notebook: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notebook deleted"})
}

// Handler для получения страниц блокнота
func getPagesHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	parts := strings.Split(urlPath, "/")

	if len(parts) < 4 {
		log.Printf("Invalid URL format: %s", urlPath)
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	notebookID, err := strconv.Atoi(parts[3])
	if err != nil {
		log.Printf("Invalid notebook ID: %s", parts[3])
		http.Error(w, "Invalid notebook ID", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching pages for notebook ID: %d", notebookID)

	pages, err := getPagesByNotebookID(notebookID)
	if err != nil {
		log.Printf("Error fetching pages for notebook ID %d: %v", notebookID, err)
		http.Error(w, fmt.Sprintf("Error fetching pages: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d pages for notebook ID %d", len(pages), notebookID)

	w.Header().Set("Content-Type", "application/json")
	if len(pages) == 0 {
		log.Printf("No pages found for notebook ID %d", notebookID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Page{}) // Возвращаем пустой массив
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pages)
	}
}

// createPageHandler — обработчик для создания страницы
func createPageHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем notebookID из URL
	notebookIDStr := r.URL.Query().Get("notebook_id")
	if notebookIDStr == "" {
		http.Error(w, "Missing notebook_id in query parameters", http.StatusBadRequest)
		return
	}

	notebookID, err := strconv.Atoi(notebookIDStr)
	if err != nil {
		http.Error(w, "Invalid notebook_id format", http.StatusBadRequest)
		return
	}

	// Парсим данные страницы из тела запроса
	var page struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&page); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Создаем структуру для страницы
	pageToInsert := Page{
		NotebookID: notebookID, // Используем полученный notebookID
		Title:      page.Title,
		Content:    page.Content,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Вставляем страницу в базу данных
	err = insertPage(pageToInsert)
	if err != nil {
		http.Error(w, "Failed to create page: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
		ID      int    `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}{
		Message: "Page created successfully",
		ID:      pageToInsert.ID,
		Title:   pageToInsert.Title,
		Content: pageToInsert.Content,
	})
}

// Обработчик для обновления блокнота
func updatePagesHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор блокнота из URL
	pageIDStr := r.URL.Path[len("/api/pages/"):]

	if pageIDStr == "" {
		http.Error(w, "Missing pages_id in URL path", http.StatusBadRequest)
		return
	}

	pageID, err := strconv.Atoi(pageIDStr)
	if err != nil {
		http.Error(w, "Invalid pages_id format", http.StatusBadRequest)
		return
	}

	// Читаем данные из тела запроса
	page := Page{}
	err = json.NewDecoder(r.Body).Decode(&page)
	page.ID = pageID
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Выполняем обновление в базе данных
	err = UpdatePage(page)
	if err != nil {
		http.Error(w, "Failed to update page: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Page updated"})
}

// Handler для удаления страницы
func deletePagesHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор блокнота из URL
	pageIDStr := r.URL.Path[len("/api/pages/"):]

	if pageIDStr == "" {
		http.Error(w, "Missing pages_id in URL path", http.StatusBadRequest)
		return
	}

	pageID, err := strconv.Atoi(pageIDStr)
	if err != nil {
		http.Error(w, "Invalid pages_id format", http.StatusBadRequest)
		return
	}

	// Читаем данные из тела запроса
	page := Page{}
	//err = json.NewDecoder(r.Body).Decode(&page)
	page.ID = pageID
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Выполняем обновление в базе данных
	err = DeletePage(page)
	if err != nil {
		http.Error(w, "Failed to update page: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Page deleted"})
}

// Handler для получения задач страницы
func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	parts := strings.Split(urlPath, "/")

	if len(parts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	pageID, err := strconv.Atoi(parts[3]) // 3-й элемент - это pageID
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching tasks for page ID: %d", pageID)

	// Получаем задачи для страницы из базы данных
	tasks, err := getTasksByPageID(pageID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching tasks: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d tasks for page ID %d", len(tasks), pageID)

	// Если задач нет, возвращаем пустой массив
	if len(tasks) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Task{}) // Пустой массив
		return
	}

	// Отправляем задачи в ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID страницы из URL
	pageIDStr := r.URL.Query().Get("page_id")
	if pageIDStr == "" {
		http.Error(w, "Page ID is required", http.StatusBadRequest)
		return
	}

	// Преобразуем pageID из строки в целое число
	pageID, err := strconv.Atoi(pageIDStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	// Декодируем тело запроса в структуру Task
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Присваиваем полученное PageID
	task.PageID = pageID

	// Вставка задачи в базу данных
	if err := insertTask(task); err != nil {
		log.Printf("Error inserting task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту с данными задачи
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func updateTasksHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор блокнота из URL
	taskIDStr := r.URL.Path[len("/api/tasks/"):]

	if taskIDStr == "" {
		http.Error(w, "Missing task_id in URL path", http.StatusBadRequest)
		return
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task_id format", http.StatusBadRequest)
		return
	}

	// Читаем данные из тела запроса
	task := Task{}
	err = json.NewDecoder(r.Body).Decode(&task)
	task.ID = taskID
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Выполняем обновление в базе данных
	err = UpdateTask(task)
	if err != nil {
		http.Error(w, "Failed to update task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Page updated"})
}

func deleteTasksHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор блокнота из URL
	taskIDStr := r.URL.Path[len("/api/tasks/"):]

	if taskIDStr == "" {
		http.Error(w, "Missing task_id in URL path", http.StatusBadRequest)
		return
	}

	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task_id format", http.StatusBadRequest)
		return
	}

	// Читаем данные из тела запроса
	task := Task{}
	//err = json.NewDecoder(r.Body).Decode(&task)
	task.ID = taskID
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Выполняем обновление в базе данных
	err = DeleteTask(task)
	if err != nil {
		http.Error(w, "Failed to delete task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted"})
}

/*func profileHandler(w http.ResponseWriter, r *http.Request) {
	// Получение токена из заголовка
	userID, err := validateAuthorization(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Извлечение данных пользователя из базы данных
	user, err := findUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Ответ с данными профиля
	response := map[string]interface{}{
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	}
	json.NewEncoder(w).Encode(response)
}
*/
