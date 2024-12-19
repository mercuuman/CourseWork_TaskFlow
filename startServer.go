package main

import (
	"log"
	"net/http"
)

func startServer() {

	if err := initDB(); err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer closeDB()

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	handleRoute(mux, "/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			mainPageHandler(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

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

	api := http.NewServeMux()

	api.HandleFunc("/api/notebooks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			updateNotebookHandler(w, r)
		} else if r.Method == http.MethodDelete {
			deleteNotebookHandler(w, r)
		}
	})

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
	// Профиль
	handleRoute(mux, "/profile", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			//ProfileGetHandler(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	apiWithAuth := tokenAuthMiddleware(api)

	mux.Handle("/api/", apiWithAuth)

	handlerWithMiddlewares := generalMiddleware(mux)

	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", handlerWithMiddlewares); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
