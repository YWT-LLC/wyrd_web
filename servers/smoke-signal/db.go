/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            TEXT PRIMARY KEY,
	username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
	password_hash TEXT NOT NULL,
	display_name  TEXT NOT NULL,
	thumbnail     TEXT,
	created_at    INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
	token_hash TEXT PRIMARY KEY,
	user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS conversations (
	id         TEXT PRIMARY KEY,
	type       TEXT NOT NULL CHECK (type IN ('bonfire','dm','group')),
	name       TEXT,
	dm_key     TEXT UNIQUE,
	created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
	created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS conversation_members (
	conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
	user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	joined_at       INTEGER NOT NULL,
	PRIMARY KEY (conversation_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_members_user ON conversation_members(user_id);

CREATE TABLE IF NOT EXISTS messages (
	seq             INTEGER PRIMARY KEY AUTOINCREMENT,
	id              TEXT NOT NULL UNIQUE,
	conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
	sender_id       TEXT REFERENCES users(id) ON DELETE SET NULL,
	content         TEXT NOT NULL,
	created_at      INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id, seq);

CREATE TABLE IF NOT EXISTS attachments (
	id          TEXT PRIMARY KEY,
	message_id  TEXT REFERENCES messages(id) ON DELETE CASCADE,
	uploader_id TEXT REFERENCES users(id) ON DELETE SET NULL,
	kind        TEXT NOT NULL CHECK (kind IN ('image','audio','video')),
	mime        TEXT NOT NULL,
	size        INTEGER NOT NULL,
	filename    TEXT NOT NULL,
	path        TEXT NOT NULL,
	created_at  INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_attachments_message ON attachments(message_id);
`

func migrate(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	// The public room everyone is in.
	_, err := db.Exec(
		`INSERT OR IGNORE INTO conversations (id, type, name, created_at) VALUES (?, 'bonfire', 'Main Bonfire', ?)`,
		mainBonfireID, time.Now().UnixMilli(),
	)
	return err
}

func newID() string { return uuid.NewString() }

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// janitor removes uploads that were never attached to a message, and expired sessions.
func (s *Server) janitor() {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	s.cleanup()
	for range t.C {
		s.cleanup()
	}
}

func (s *Server) cleanup() {
	now := time.Now()
	cutoff := now.Add(-24 * time.Hour).UnixMilli()

	rows, err := s.db.Query(
		`SELECT id, path FROM attachments WHERE message_id IS NULL AND created_at < ?`, cutoff)
	if err != nil {
		log.Printf("cleanup query: %v", err)
		return
	}
	type stale struct{ id, path string }
	var list []stale
	for rows.Next() {
		var st stale
		if rows.Scan(&st.id, &st.path) == nil {
			list = append(list, st)
		}
	}
	rows.Close()

	for _, st := range list {
		os.Remove(filepath.Join(s.dataDir, "uploads", st.path))
		s.db.Exec(`DELETE FROM attachments WHERE id = ?`, st.id)
	}
	s.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, now.UnixMilli())
}