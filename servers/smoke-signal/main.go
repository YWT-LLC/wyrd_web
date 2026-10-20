/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

//* Define the build data *//

const (
	maxThumbnailBytes = 5 << 20  // 5 MB
	maxMediaBytes     = 75 << 20 // 75 MB
	mainBonfireID     = "main-bonfire"
)

type Server struct {
	db         *sql.DB
	hub        *Hub
	baseUrl    string
	dataDir    string
	inviteCode string
}

//* Define custom functions *//

// Attempt to retrieve env var, return fallback to sender on "failure"
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dataDir := envOr("SMOKE_DATA_DIR", "./data")
	for _, sub := range []string{"uploads", "thumbnails"} {
		if err := os.MkdirAll(filepath.Join(dataDir, sub), 0o755); err != nil {
			log.Fatal(err)
		}
	}

	// Pragmas in the DSN apply to every pooled connection.
	dsn := "file:" + filepath.Join(dataDir, "smoke.db") +
		"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		log.Fatal(err)
	}

	s := &Server{
		db:         db,
		hub:        newHub(),
		baseUrl:    envOr("SMOKE_BASE_URL", "http://localhost:8080"),
		dataDir:    dataDir,
		inviteCode: os.Getenv("SMOKE_INVITE_CODE"),
	}
	go s.janitor()

	addr := envOr("SMOKE_ADDR", ":8080")
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.routes(),
		ReadHeaderTimeout: 10 * time.Second,
		// No Read/WriteTimeout: they would kill large uploads and WebSockets.
	}
	log.Printf("Smoke Signal listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Accounts
	mux.HandleFunc("POST /api/register", s.handleRegister)
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.auth(s.handleLogout))
	mux.HandleFunc("GET /api/me", s.auth(s.handleGetMe))
	mux.HandleFunc("PATCH /api/me", s.auth(s.handleUpdateMe))
	mux.HandleFunc("DELETE /api/me", s.auth(s.handleDeleteMe))
	mux.HandleFunc("PUT /api/me/thumbnail", s.auth(s.handleSetThumbnail))
	mux.HandleFunc("DELETE /api/me/thumbnail", s.auth(s.handleDeleteThumbnail))
	mux.HandleFunc("GET /api/users", s.auth(s.handleSearchUsers))
	mux.HandleFunc("GET /api/users/{id}", s.auth(s.handleGetUser))
	mux.HandleFunc("GET /api/users/{id}/thumbnail", s.auth(s.handleGetThumbnail))

	// Conversations (bonfire / dm / group)
	mux.HandleFunc("GET /api/conversations", s.auth(s.handleListConversations))
	mux.HandleFunc("POST /api/conversations/dm", s.auth(s.handleCreateDM))
	mux.HandleFunc("POST /api/conversations/group", s.auth(s.handleCreateGroup))
	mux.HandleFunc("POST /api/conversations/{id}/members", s.auth(s.handleAddMember))
	mux.HandleFunc("DELETE /api/conversations/{id}/members/me", s.auth(s.handleLeaveGroup))
	mux.HandleFunc("GET /api/conversations/{id}/messages", s.auth(s.handleListMessages))

	// Media
	mux.HandleFunc("POST /api/uploads", s.auth(s.handleUpload))
	mux.HandleFunc("GET /api/files/{id}", s.auth(s.handleGetFile))

	// Realtime
	mux.HandleFunc("GET /ws", s.handleWS)

	return mux
}
