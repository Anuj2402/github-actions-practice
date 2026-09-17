package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Message struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

var messages = []Message{
	{ID: 1, User: "system", Message: "Welcome to the chat app!", CreatedAt: time.Now()},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/history", serveHistory)
	mux.HandleFunc("/api/chat-history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(messages)
	})

	mux.HandleFunc("/api/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			User    string `json:"user"`
			Message string `json:"message"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if payload.User == "" || payload.Message == "" {
			http.Error(w, "user and message are required", http.StatusBadRequest)
			return
		}

		messages = append(messages, Message{
			ID:        len(messages) + 1,
			User:      payload.User,
			Message:   payload.Message,
			CreatedAt: time.Now(),
		})

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	fmt.Println("Chat app listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func serveHistory(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "history.html")
}
