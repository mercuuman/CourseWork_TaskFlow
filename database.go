package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
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

// Функция для поиска пользователя в бд
func findUser(user *User) (bool, int, error) {
	// Выполняем запрос, чтобы получить id пользователя, если существует такая пара username и password
	query := "SELECT id FROM users WHERE username = $1 AND password = $2 LIMIT 1"
	var userID int

	// Если пользователь найден, возвращаем его id и состояние существования
	err := db.QueryRow(context.Background(), query, user.Username, user.Password).Scan(&userID)
	if err != nil {
		// Если пользователь не найден, возвращаем false
		if err == sql.ErrNoRows {
			return false, 0, nil // Пользователь не найден
		}
		// Если ошибка другая, возвращаем её
		return false, 0, err
	}

	// Если мы получаем id, значит, пользователь существует
	return true, userID, nil
}
func getUserFromDB(id int) (User, error) {
	var user User
	query := "SELECT id, username, password, email,created_at FROM users WHERE id = $1"

	err := db.QueryRow(context.Background(), query, id).Scan(user.ID, user.Username, user.Email, user.Password, user.CreatedAt)

	if err != nil {
		return user, err
	}
	return user, nil
}

// Вывод блокнотов пользователя
func getNotebooksByUserID(userID int) ([]Notebook, error) {
	query := "SELECT id, user_id, name, created_at, updated_at FROM notebooks WHERE user_id = $1"
	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении блокнотов: %v", err)
	}
	defer rows.Close()

	var notebooks []Notebook
	for rows.Next() {
		var notebook Notebook
		if err := rows.Scan(&notebook.ID, &notebook.UserID, &notebook.Name, &notebook.CreatedAt, &notebook.UpdatedAt); err != nil {
			return nil, fmt.Errorf("Ошибка при сканировании данных блокнота: %v", err)
		}
		notebooks = append(notebooks, notebook)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Ошибка при обработке результатов запроса: %v", err)
	}

	return notebooks, nil
}

// Вставка блокнота
func insertNotebook(notebook Notebook) error {
	// Логирование данных перед вставкой
	log.Printf("Inserting notebook into DB: %+v", notebook)

	// Создаем SQL запрос для вставки блокнота
	query := "INSERT INTO notebooks (user_id, name) VALUES ($1, $2)"
	_, err := db.Exec(context.Background(), query, notebook.UserID, notebook.Name)

	if err != nil {
		return fmt.Errorf("Ошибка при добавлении блокнота: %v", err)
	}
	log.Println("Notebook inserted into DB successfully")
	return nil
}

func UpdateNotebook(notebook Notebook) error {
	log.Printf("Updating notebook into DB: %+v", notebook)
	query := "UPDATE notebooks SET name = $1 WHERE id = $2"
	_, err := db.Exec(context.Background(), query, notebook.Name, notebook.ID)
	if err != nil {
		return fmt.Errorf("Ошибка при обнолвении блокнота: %v", err)
	}
	log.Println("Notebook updated into DB successfully")
	return nil
}

func DeleteNotebook(notebook Notebook) error {
	log.Printf("Deleting notebook from DB: %+v", notebook)

	// Используем DELETE вместо UPDATE
	query := "DELETE FROM notebooks WHERE id = $1"

	// Выполняем запрос удаления
	_, err := db.Exec(context.Background(), query, notebook.ID)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении блокнота: %v", err)
	}

	log.Println("Notebook deleted from DB successfully")
	return nil
}

// Вывод страниц из блокнота
func getPagesByNotebookID(notebookID int) ([]Page, error) {
	query := "SELECT id, title,content, created_at, updated_at FROM pages WHERE notebook_id = $1"
	rows, err := db.Query(context.Background(), query, notebookID)
	if err != nil {
		return nil, fmt.Errorf("Error fetching pages: %v", err)
	}
	defer rows.Close()

	var pages []Page
	for rows.Next() {
		var page Page
		if err := rows.Scan(&page.ID, &page.Title, &page.Content, &page.CreatedAt, &page.UpdatedAt); err != nil {
			return nil, fmt.Errorf("Error scanning page data: %v", err)
		}
		pages = append(pages, page)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating over rows: %v", err)
	}

	// Логируем количество страниц и данные
	log.Printf("Found %d pages for notebook ID %d", len(pages), notebookID)
	for _, page := range pages {
		log.Printf("Page ID: %d, Title: %s, CreatedAt: %s, UpdatedAt: %s", page.ID, page.Title, page.CreatedAt, page.UpdatedAt)
	}

	return pages, nil
}

// Вставка страницы
func insertPage(page Page) error {
	// Логирование данных перед вставкой
	log.Printf("Inserting page into DB: %+v", page)

	// Создаем SQL запрос для вставки страницы
	query := "INSERT INTO pages (notebook_id, title, content) VALUES ($1, $2, $3)"
	_, err := db.Exec(context.Background(), query, page.NotebookID, page.Title, page.Content)

	if err != nil {
		return fmt.Errorf("Ошибка при добавлении страницы: %v", err)
	}
	log.Println("Page inserted into DB successfully")
	return nil
}

func UpdatePage(page Page) error {
	log.Printf("Updating page into DB: %+v", page)
	query := "UPDATE pages SET title = $1,content=$2 WHERE id = $3"
	_, err := db.Exec(context.Background(), query, page.Title, page.Content, page.ID)
	if err != nil {
		return fmt.Errorf("Ошибка при обновлении страницы: %v", err)
	}
	log.Println("Page updated into DB successfully")
	return nil
}

func DeletePage(page Page) error {
	log.Printf("Deleting page from DB: %+v", page)

	query := "DELETE FROM pages WHERE id = $1"

	// Выполняем запрос удаления
	_, err := db.Exec(context.Background(), query, page.ID)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении страницы: %v", err)
	}

	log.Println("page deleted from DB successfully")
	return nil
}

//Вывод задач страницы

func getTasksByPageID(pageID int) ([]Task, error) {
	query := "SELECT id, page_id, title, description, status, priority, due_date, created_at, updated_at FROM tasks WHERE page_id = $1"
	rows, err := db.Query(context.Background(), query, pageID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении задач: %v", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.PageID, &task.Title, &task.Description, &task.Status, &task.Priority, &task.DueDate, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, fmt.Errorf("Ошибка при сканировании данных задачи: %v", err)
		}
		tasks = append(tasks, task)
	}

	// Логируем количество извлеченных задач
	log.Printf("Found %d tasks for page ID %d", len(tasks), pageID)

	if len(tasks) == 0 {
		log.Println("No tasks found for page ID", pageID)
	}

	return tasks, nil
}

// Функция для вставки новой задачи
func insertTask(task Task) error {
	// Логирование данных перед вставкой
	log.Printf("Inserting task into DB: %+v", task)

	// Если due_date пустое, присваиваем текущую дату или оставляем NULL
	if (task.DueDate == time.Time{}) { // Если пустое значение time.Time
		task.DueDate = time.Now() // Устанавливаем текущую дату и время
	}

	// Создаем SQL запрос для вставки задачи
	query := "INSERT INTO tasks (page_id, title, description, status, priority, due_date) VALUES ($1, $2, $3, $4, $5, $6)"

	// Выполняем SQL запрос
	_, err := db.Exec(context.Background(), query, task.PageID, task.Title, task.Description, task.Status, task.Priority, task.DueDate)

	if err != nil {
		// Если произошла ошибка, логируем и возвращаем ошибку
		return fmt.Errorf("Ошибка при добавлении задачи: %v", err)
	}

	// Логируем успешную вставку
	log.Println("Task inserted into DB successfully")
	return nil
}

func UpdateTask(task Task) error {
	log.Printf("Updating task into DB: %+v", task)
	query := "UPDATE tasks SET title = $1,description=$2 WHERE id = $3"
	_, err := db.Exec(context.Background(), query, task.Title, task.Description, task.ID)
	if err != nil {
		return fmt.Errorf("Ошибка при обновлении задачи: %v", err)
	}
	log.Println("Task updated into DB successfully")
	return nil
}

func DeleteTask(task Task) error {
	log.Printf("Deleting task from DB: %+v", task)

	query := "DELETE FROM tasks WHERE id = $1"

	// Выполняем запрос удаления
	_, err := db.Exec(context.Background(), query, task.ID)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении Задачи: %v", err)
	}

	log.Println("task deleted from DB successfully")
	return nil
}
