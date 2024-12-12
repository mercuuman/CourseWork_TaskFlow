package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	accessSecret    = []byte("TaskMasterTaskFlowCourseWork") // Секретный ключ для подписи токенов
	refreshSecret   = []byte("7daysitsonlytimewhenwelive")
	accessTokenTTL  = time.Minute * 15   // Время жизни токена доступа
	refreshTokenTTL = time.Hour * 24 * 7 // Время жизни токена обновления
)

// Middleware для проверки JWT
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
