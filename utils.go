package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	accessSecret    = []byte("TaskMasterTaskFlowCourseWork") // Секретный ключ для подписи токенов
	refreshSecret   = []byte("7daysitsonlytimewhenwelive")
	accessTokenTTL  = time.Minute * 15   // Время жизни токена доступа
	refreshTokenTTL = time.Hour * 24 * 7 // Время жизни токена обновления
)

// Middleware для обработки CORS и логирования
func generalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Логируем запрос
		log.Printf("Request received: %s %s", r.Method, r.URL.Path)

		// CORS Headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Если это OPTIONS запрос, сразу отвечаем
		if r.Method == http.MethodOptions {
			log.Printf("OPTIONS request for %s", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}

// Универсальная функция для обработки маршрутов
func handleRoute(mux *http.ServeMux, pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	mux.HandleFunc(pattern, handlerFunc)
}

// Middleware для проверки JWT УДАЛИТЬ???
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Получение заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// 2. Извлечение токена (удаление префикса "Bearer ")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. Валидация токена
		claims, err := validateToken(tokenString, accessSecret)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// 4. Токен валиден, передача управления следующему обработчику
		fmt.Println("Authenticated user ID:", claims["userID"])
		next.ServeHTTP(w, r)
	})
}

func tokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовков
		token := r.Header.Get("Authorization")
		if token == "" {
			// Если токен отсутствует, перенаправляем на страницу логина
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Убираем префикс "Bearer" из токена, если он есть
		token = strings.TrimPrefix(token, "Bearer ")

		// Проверяем токен
		claims, err := validateToken(token, accessSecret)
		if err != nil {
			// Если токен недействителен, перенаправляем на страницу логина
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Добавляем информацию о пользователе в контекст
		userID, ok := claims["userID"].(string)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Добавляем userID в контекст запроса
		ctx := context.WithValue(r.Context(), "userID", userID)

		// Вызываем следующий обработчик с обновленным контекстом
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateToken(tokenString string, secretKey []byte) (jwt.MapClaims, error) {
	// 1. Разбор токена
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// 1.1 Проверка метода подписи
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		// 1.2 Возврат секретного ключа для проверки подписи
		return secretKey, nil
	})
	if err != nil {
		return nil, err // Ошибка парсинга токена
	}

	// 2. Извлечение claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// 3. Проверка срока действия токена
	if exp, ok := claims["ExpiresAt"].(float64); ok {
		// Если срок действия истёк, возвращаем ошибку
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return nil, fmt.Errorf("token expired")
		}
	} else {
		return nil, fmt.Errorf("invalid expiration field")
	}

	return claims, nil // Токен валиден, возвращаем claims
}

func generateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"userID":    userID,
		"ExpiresAt": time.Now().Add(accessTokenTTL).Unix(), // Срок действия 15 минут
		"IssuedAt":  time.Now(),                            // Время создания токена
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // Создание токена с алгоритмом HS256
	return token.SignedString(accessSecret)                    // Подписание токена другим секретным ключом
}

func generateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"userID":    userID,
		"ExpiresAt": time.Now().Add(refreshTokenTTL).Unix(), // Срок действия 7 дней
		"IssuedAt":  time.Now(),                             // Время создания токена
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecret)
}

func handleError(w http.ResponseWriter, err error, status int) {
	log.Println("Error:", err) // Логирование ошибки
	w.WriteHeader(status)
	response := map[string]string{"error": err.Error()}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getUserIDFromToken(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, fmt.Errorf("заголовок Authorization отсутствует")
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return 0, fmt.Errorf("заголовок Authorization некорректен")
	}

	tokenString := authHeader[7:]
	log.Println("Extracted token string:", tokenString)

	claims, err := validateToken(tokenString, accessSecret)
	if err != nil {
		return 0, fmt.Errorf("неверный токен: %v", err)
	}

	userIDStr, ok := claims["userID"].(string)
	if !ok {
		return 0, fmt.Errorf("userID отсутствует или некорректен: %v", claims["userID"])
	}

	log.Println("Extracted userID from claims:", userIDStr)

	// Преобразование в int
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("ошибка преобразования userID в int: %v", err)
	}

	return userID, nil
}
