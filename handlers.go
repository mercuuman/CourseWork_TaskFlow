package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

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

	accessToken, err := generateAccessToken(string(userID))
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	refreshToken, err := generateRefreshToken(string(userID))
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"accessToken": "%s"}`, accessToken)
}

func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получение refresh-токена из запроса
	var requestBody struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	refreshToken := requestBody.RefreshToken

	// 2. Проверка refresh-токена
	claims, err := validateToken(refreshToken, refreshSecret)
	if err != nil {
		http.Error(w, "Invalid refresh token: "+err.Error(), http.StatusUnauthorized)
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

	// 5. (Опционально) Генерация нового refresh-токена
	newRefreshToken, err := generateRefreshToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// 6. Отправка новых токенов клиенту
	response := struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

/*
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
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
}*/
