package main

// Структура пользователя
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Структура для обработки данных регистрации
type SignUpRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	//ConfirmPassword string `json:"confirm_password"`
}

// Структура для обработки данных авторизации
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Структура для создания блокнота
type Notebook struct {
	UserID    int    `json:"user_id"`
	Name      string `json:"name"`
	UpdatedAt string `json:"updated_at"`
}
