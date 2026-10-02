package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// Define structs //

type Client struct {
	conn *websocket.Conn
}

type Bonfire struct {
	clients   map[*Client]bool
	broadcast chan Message
	mu        sync.Mutex
}

type Message struct {
	ID        string    `json:"id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Define helpers //

func startBonfire() *Bonfire {
	return &Bonfire{
		clients:   make(map[*Client]bool),
		broadcast: make(chan Message),
	}
}

func (h *Bonfire) run(db *sql.DB) {
	for msg := range h.broadcast {
		// Persist to SQLite
		_, err := db.Exec(
			"INSERT INTO messages (id, sender, content, created_at) VALUES (?, ?, ?, ?)",
			msg.ID, msg.Sender, msg.Content, msg.CreatedAt,
		)
		if err != nil {
			log.Printf("Database write error: %v", err)
			continue
		}

		// Broadcast to all active connected clients
		data, _ := json.Marshal(msg)
		h.mu.Lock()
		for client := range h.clients {
			err := client.conn.Write(context.Background(), websocket.MessageText, data)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.conn.Close(websocket.StatusInternalError, "write failed")
				delete(h.clients, client)
			}
		}
		h.mu.Unlock()
	}
}

// Make it so //

func main() {
	// Initialize SQLite
	db, err := sql.Open("sqlite", "./chat.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ActivityPub-friendly Schema (for later)
	schema := `
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		sender TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`
	if _, err := db.Exec(schema); err != nil {
		log.Fatal(err)
	}

	bonfire := startBonfire()
	go bonfire.run(db)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // Allows connections from emulator
		})
		if err != nil {
			log.Println("Connection failed:", err)
			return
		}

		client := &Client{conn: conn}
		bonfire.mu.Lock()
		bonfire.clients[client] = true
		bonfire.mu.Unlock()

		defer func() {
			bonfire.mu.Lock()
			delete(bonfire.clients, client)
			bonfire.mu.Unlock()
			conn.Close(websocket.StatusNormalClosure, "")
		}()

		// Read loop
		for {
			var msg Message
			_, data, err := conn.Read(r.Context())
			if err != nil {
				break
			}
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}

			// Assign server-side URI ID and timestamp
			msg.ID = fmt.Sprintf("https://localhost/notes/%s", uuid.New().String())
			msg.CreatedAt = time.Now().UTC()

			bonfire.broadcast <- msg
		}
	})

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}