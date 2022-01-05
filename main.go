package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// Dependency stands for dependency actually. I made this up.
type Dependency struct {
	DB *sql.DB
}

func main() {
	dbURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer db.Close()

	deps := Dependency{DB: db}

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  err = deps.Migrate(ctx)
  if err != nil {
    log.Fatalf("migrating: %v", err)
  }

	r := http.NewServeMux()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	r.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if id := r.URL.Query().Get("id"); id != "" {
				idInt, err := strconv.Atoi(id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				if post, err := deps.GetPostById(r.Context(), idInt); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(post)
				}
			} else {
				if posts, err := deps.GetPosts(r.Context()); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(posts)
				}
			}

		case http.MethodPost:
			var post Post
			if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if post, err := deps.AddPost(r.Context(), post); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(post)
			}

		case http.MethodPatch:
			var post Post
			if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if post, err := deps.AddPost(r.Context(), post); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(post)
			}

		case http.MethodDelete:
			if id := r.URL.Query().Get("id"); id != "" {
				idInt, err := strconv.Atoi(id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				if _, err := deps.GetPostById(r.Context(), idInt); err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
			} else {
				http.Error(w, "id is required", http.StatusBadRequest)
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	r.HandleFunc("/comment", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if id := r.URL.Query().Get("id"); id != "" {
				idInt, err := strconv.Atoi(id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				if comment, err := deps.GetComments(r.Context(), idInt); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(comment)
				}
			} else {
				http.Error(w, "id is required", http.StatusBadRequest)
				return
			}

		case http.MethodPost:
			var comment Comment
			if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if comment, err := deps.AddComment(r.Context(), comment); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(comment)
			}

		case http.MethodDelete:
			if id := r.URL.Query().Get("id"); id != "" {
				idInt, err := strconv.Atoi(id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				if _, err := deps.DeleteComment(r.Context(), idInt); err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
			} else {
				http.Error(w, "id is required", http.StatusBadRequest)
				return
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	<-sigCh
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server gracefully stopped")
}
