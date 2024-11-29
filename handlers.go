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
