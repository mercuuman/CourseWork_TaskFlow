package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

var (
	db              *pgxpool.Pool
	ErrDuplicateKey = errors.New("duplicate key") // Экспортируемая ошибка
)

// Инициализация базы данных
func initDB() error {
	connStr := "postgresql://postgres:641@localhost:5432/task_management" // Укажите свои данные
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}

	db = pool
	return nil
}

// Закрытие подключения к базе данных
func closeDB() {
	if db != nil {
		db.Close()
	}
}

// Функция для добавления пользователя в базу данных
func insertUser(user User) error {
	// Логирование данных перед вставкой
	log.Printf("Inserting user into DB: %+v", user)

	// Создаем SQL запрос для вставки пользователя
	query := "INSERT INTO users (email, username, password) VALUES ($1, $2, $3)"
	_, err := db.Exec(context.Background(), query, user.Email, user.Username, user.Password)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Код SQLSTATE для уникального ограничения
			switch pgErr.ConstraintName {
			case "users_email_key": // Уникальное ограничение на email
				return fmt.Errorf("email already exists: %w", ErrDuplicateKey)
			case "users_username_key": // Уникальное ограничение на username
				return fmt.Errorf("username already exists: %w", ErrDuplicateKey)
			}
		}
		return fmt.Errorf("Ошибка при добавлении пользователя: %v", err)
	}
	log.Println("User inserted into DB successfully")
	return nil
}
