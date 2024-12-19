package main

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

// Mock пользователь для тестирования
var testUser = User{
	Username: "test_userrr",
	Email:    "test_userrr@example.com",
	Password: "password123",
}
var testNotebook = Notebook{
	UserID: 6,
	Name:   "Test Notebook",
}
var testPage = Page{
	NotebookID: 8,
	Title:      "Test Page",
	Content:    "Test Content",
}
var testTask = Task{
	PageID:      26,
	Title:       "Test Task",
	Description: "Test Description",
}

func TestInsertUser(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	err := insertUser(testUser)
	if err != nil {
		assert.Contains(t, err.Error(), "duplicate key", "Ошибка должна быть связана с уникальностью ключа, если пользователь уже существует")
	} else {
		assert.NoError(t, err, "Пользователь должен быть успешно добавлен")
	}
}

/*func TestGetUserFromDB(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	_, userID, err := findUser(&testUser)
	require.NoError(t, err)

	user, err := getUserFromDB(userID)
	assert.NoError(t, err, "Получение пользователя из базы данных не должно вызывать ошибку")
	assert.Equal(t, testUser.Username, user.Username, "Имена пользователей должны совпадать")
}*/

func TestInsertNotebook(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	err := insertNotebook(testNotebook)
	assert.NoError(t, err, "Блокнот должен быть успешно добавлен")
}

func TestGetNotebooksByUserID(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	notebooks, err := getNotebooksByUserID(testNotebook.UserID)
	assert.NoError(t, err, "Получение блокнотов пользователя не должно возвращать ошибку")
	assert.NotEmpty(t, notebooks, "Список блокнотов не должен быть пустым")
}

func TestUpdateNotebook(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	testNotebook.Name = "Updated Notebook"
	err := UpdateNotebook(testNotebook)
	assert.NoError(t, err, "Обновление блокнота не должно вызывать ошибку")
}

func TestDeleteNotebook(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	err := DeleteNotebook(testNotebook)
	assert.NoError(t, err, "Удаление блокнота не должно вызывать ошибку")
}

func TestInsertPage(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	err := insertPage(testPage)
	assert.NoError(t, err, "Страница должна быть успешно добавлена")
}

func TestGetPagesByNotebookID(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	pages, err := getPagesByNotebookID(testPage.NotebookID)
	assert.NoError(t, err, "Получение страниц блокнота не должно возвращать ошибку")
	assert.NotEmpty(t, pages, "Список страниц не должен быть пустым")
}

func TestUpdatePage(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	testPage.Title = "Updated Page"
	err := UpdatePage(testPage)
	assert.NoError(t, err, "Обновление страницы не должно вызывать ошибку")
}

func TestDeletePage(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	err := DeletePage(testPage)
	assert.NoError(t, err, "Удаление страницы не должно вызывать ошибку")
}

func TestInsertTask(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	err := insertTask(testTask)
	assert.NoError(t, err, "Задача должна быть успешно добавлена")
}

func TestGetTasksByPageID(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	tasks, err := getTasksByPageID(testTask.PageID)
	assert.NoError(t, err, "Получение задач страницы не должно возвращать ошибку")
	assert.NotEmpty(t, tasks, "Список задач не должен быть пустым")
}

func TestUpdateTask(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	testTask.Title = "Updated Task"
	err := UpdateTask(testTask)
	assert.NoError(t, err, "Обновление задачи не должно вызывать ошибку")
}

func TestDeleteTask(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	err := DeleteTask(testTask)
	assert.NoError(t, err, "Удаление задачи не должно вызывать ошибку")
}
