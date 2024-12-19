package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFeHBpcmVzQXQiOjE3MzQ2MzA3NzAsIklzc3VlZEF0IjoiMjAyNC0xMi0xOVQyMTozNzo1MC40Nzg1OTAxKzA0OjAwIiwidXNlcklEIjoiNjQifQ.-A7teO0h5HLIGtDu7B6Z-_1MjjCYSjs4QsDKrdgwQEQ"

// Тест главной страницы
/*func TestMainPageHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(mainPageHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "<html>") {
		t.Errorf("handler returned unexpected body: %v", rr.Body.String())
	}
}
*/
// Тест на регистрацию пользователя (POST /signup)
func TestSignUpPostHandler(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	reqBody := SignUpRequest{
		Email:    "test1@example.com",
		Username: "test1user",
		Password: "testpassword",
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/signup", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SignUpPostHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "User registered successfully"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// Тест на авторизацию пользователя (POST /login)
func TestLoginPostHandler(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	reqBody := LoginRequest{
		Username: "testuser",
		Password: "testpassword",
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/login", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(LoginPostHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "accessToken") {
		t.Errorf("handler returned unexpected body: %v", rr.Body.String())
	}
}

// Тест на получение блокнотов (GET /api/notebooks)
func TestGetNotebooksHandler(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	req, err := http.NewRequest("GET", "/api/notebooks", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", token)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getNotebooksHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

// Тест на создание блокнота (POST /api/notebooks)
func TestCreateNotebookHandler(t *testing.T) {
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()
	reqBody := struct {
		Name string `json:"name"`
	}{
		Name: "New Notebook",
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/api/notebooks", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", token)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createNotebookHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if !strings.Contains(rr.Body.String(), "Notebook created successfully") {
		t.Errorf("handler returned unexpected body: %v", rr.Body.String())
	}
}
