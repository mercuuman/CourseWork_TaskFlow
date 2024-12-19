package main

import (
	"log"
	"net/http"
	//"fmt"
	//"html/template"
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

func main() {
	// Инициализация базы данных
	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()

	// Создаем новый мультиплексор
	mux := http.NewServeMux()

	// Статические файлы (CSS, JS)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	handleRoute(mux, "/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			mainPageHandler(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Логика для регистрации и авторизации
	handleRoute(mux, "/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			LogInGetHandler(w, r)
		case http.MethodPost:
			LoginPostHandler(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	handleRoute(mux, "/signup", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			SignUpGetHandler(w, r)
		case http.MethodPost:
			SignUpPostHandler(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Профиль
	handleRoute(mux, "/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			//ProfileGetHandler(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Защищенные API маршруты для блокнотов, страниц и задач
	api := http.NewServeMux()

	// Маршрут для PUT-запроса на обновление блокнота
	api.HandleFunc("/api/notebooks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			updateNotebookHandler(w, r)
		} else if r.Method == http.MethodDelete {
			deleteNotebookHandler(w, r)
		}
	})

	// Применяем токен-авторизацию вручную в каждом обработчике
	api.HandleFunc("/api/notebooks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getNotebooksHandler(w, r)
		} else if r.Method == http.MethodPost {
			createNotebookHandler(w, r)
		} else if r.Method == http.MethodPut {
			updateNotebookHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	api.HandleFunc("/api/pages/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getPagesHandler(w, r)
		} else if r.Method == http.MethodPost {
			createPageHandler(w, r)
		} else if r.Method == http.MethodPut {
			updatePagesHandler(w, r)
		} else if r.Method == http.MethodDelete {
			deletePagesHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	/*	api.HandleFunc("/refresh-token", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				log.Printf("POST /refresh-token")
				refreshTokenHandler(w, r)
			}
		})
	*/

	mux.HandleFunc("/refresh-token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			log.Printf("POST /refresh-token")
			refreshTokenHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	api.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getTasksHandler(w, r)
		} else if r.Method == http.MethodPost {
			createTaskHandler(w, r)
		} else if r.Method == http.MethodPut {
			updateTasksHandler(w, r)
		} else if r.Method == http.MethodDelete {
			deleteTasksHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Все маршруты API защищены авторизацией
	apiWithAuth := tokenAuthMiddleware(api)

	// Включаем API маршруты в главный мультиплексор с авторизацией
	mux.Handle("/api/", apiWithAuth)

	// Применяем общий middleware для всех маршрутов
	handlerWithMiddlewares := generalMiddleware(mux)

	// Запуск сервера
	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", handlerWithMiddlewares); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
