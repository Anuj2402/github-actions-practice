package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var db *gorm.DB

type ClientKey struct {
	From string
	To   string
}

var clients = make(map[ClientKey]*websocket.Conn)

type Message struct {
	ID         uint      `gorm:"primaryKey"`
	Datetime   time.Time `gorm:"autoCreateTime"`
	FromUser   string
	ToUser     string
	MsgContent string
}

type ChatHistoryResponse struct {
	Messages []ChatHistoryMessage `json:"messages"`
}

type ChatHistoryMessage struct {
	Timestamp string `json:"timestamp"`
	FromUser  string `json:"from"`
	ToUser    string `json:"to"`
	Content   string `json:"content"`
}

// getEnv returns the value of key if set, otherwise fallback.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func main() {
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "chatdb")
	appPort := getEnv("APP_PORT", "8080")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(" Failed to connect to DB:", err)
	}

	db.AutoMigrate(&Message{})

	http.HandleFunc("/ws", handleConnection)

	http.HandleFunc("/api/chat-history", handleChatHistory)

	// Serve static files (for the chat interface)
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Printf("Chat server started at http://localhost:%s\n", appPort)
	fmt.Printf("WebSocket endpoint: ws://localhost:%s/ws\n", appPort)
	log.Fatal(http.ListenAndServe(":"+appPort, nil))
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	fromUser := r.URL.Query().Get("from")
	toUser := r.URL.Query().Get("to")

	connKey := ClientKey{From: fromUser, To: toUser}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(" Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Printf(" %s connected to chat with %s\n", fromUser, toUser)

	clients[connKey] = conn
	defer delete(clients, connKey)

	sendChatHistory(conn, fromUser, toUser)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf(" %s disconnected from chat with %s\n", fromUser, toUser)
			break
		}

		currentTime := time.Now()
		timestamp := currentTime.Format("2006-01-02 15:04")

		message := Message{
			FromUser:   fromUser,
			ToUser:     toUser,
			MsgContent: string(msg),
			Datetime:   currentTime,
		}
		db.Create(&message)

		formatted := fmt.Sprintf("[%s] %s → %s: %s", timestamp, fromUser, toUser, msg)

		conn.WriteMessage(websocket.TextMessage, []byte(formatted))

		// Send to recipient only in chats where To=A and From=C
		for key, toConn := range clients {
			if key.From == toUser && key.To == fromUser {
				toConn.WriteMessage(websocket.TextMessage, []byte(formatted))
			}
		}
	}
}

func sendChatHistory(conn *websocket.Conn, user1, user2 string) {
	var history []Message
	db.
		Where("(from_user = ? AND to_user = ?) OR (from_user = ? AND to_user = ?)",
			user1, user2, user2, user1).
		Order("datetime").
		Find(&history)

	for _, msg := range history {
		timestamp := msg.Datetime.Format("2006-01-02 15:04")
		formatted := fmt.Sprintf("[%s] %s → %s: %s", timestamp, msg.FromUser, msg.ToUser, msg.MsgContent)
		conn.WriteMessage(websocket.TextMessage, []byte(formatted))
	}
}

func handleChatHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fromUser := r.URL.Query().Get("from")
	toUser := r.URL.Query().Get("to")
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	if fromUser == "" || toUser == "" {
		http.Error(w, "Missing required parameters: from and to", http.StatusBadRequest)
		return
	}

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse("2006-01-02 15:04:05", startTimeStr)
		if err != nil {
			http.Error(w, "Invalid start_time format. Use YYYY-MM-DD HH:MM:SS", http.StatusBadRequest)
			return
		}
	} else {
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse("2006-01-02 15:04:05", endTimeStr)
		if err != nil {
			http.Error(w, "Invalid end_time format. Use YYYY-MM-DD HH:MM:SS", http.StatusBadRequest)
			return
		}
	} else {
		endTime = time.Now()
	}

	var messages []Message
	result := db.Where("((from_user = ? AND to_user = ?) OR (from_user = ? AND to_user = ?)) AND datetime BETWEEN ? AND ?",
		fromUser, toUser, toUser, fromUser, startTime, endTime).
		Order("datetime").
		Find(&messages)

	if result.Error != nil {
		http.Error(w, "Failed to retrieve chat history: "+result.Error.Error(), http.StatusInternalServerError)
		return
	}

	response := ChatHistoryResponse{
		Messages: make([]ChatHistoryMessage, 0, len(messages)),
	}

	for _, msg := range messages {
		response.Messages = append(response.Messages, ChatHistoryMessage{
			Timestamp: msg.Datetime.Format("2006-01-02 15:04:05"),
			FromUser:  msg.FromUser,
			ToUser:    msg.ToUser,
			Content:   msg.MsgContent,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
