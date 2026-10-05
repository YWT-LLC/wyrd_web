/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package main

import (
	"database/sql"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Always query as "FROM users u".
const userCols = `u.id, u.username, u.display_name, u.thumbnail, u.created_at`

type rowScanner interface{ Scan(dest ...any) error }

func (u *User) finish(thumb sql.NullString, created int64) {
	u.CreatedAt = time.UnixMilli(created).UTC()
	if thumb.Valid {
		// The filename changes on every upload, so it doubles as a cache-buster.
		u.ThumbnailURL = "/api/users/" + u.ID + "/thumbnail?v=" + thumb.String
	}
}

func scanUser(row rowScanner) (*User, error) {
	var u User
	var thumb sql.NullString
	var created int64
	if err := row.Scan(&u.ID, &u.Username, &u.DisplayName, &thumb, &created); err != nil {
		return nil, err
	}
	u.finish(thumb, created)
	return &u, nil
}

func (s *Server) getUser(id string) (*User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userCols+` FROM users u WHERE u.id = ?`, id))
}

func validDisplayName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	n := utf8.RuneCountInString(name)
	return name, n >= 1 && n <= 64
}

func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request, uid string) {
	user, err := s.getUser(uid)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request, _ string) {
	user, err := s.getUser(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleUpdateMe(w http.ResponseWriter, r *http.Request, uid string) {
	var in struct {
		DisplayName string `json:"display_name"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	name, ok := validDisplayName(in.DisplayName)
	if !ok {
		writeError(w, http.StatusBadRequest, "display name must be 1-64 characters")
		return
	}
	if _, err := s.db.Exec(`UPDATE users SET display_name = ? WHERE id = ?`, name, uid); err != nil {
		writeError(w, http.StatusInternalServerError, "update failed")
		return
	}
	user, _ := s.getUser(uid)
	writeJSON(w, http.StatusOK, user)
}

// DELETE /api/me  {"password": "..."}
func (s *Server) handleDeleteMe(w http.ResponseWriter, r *http.Request, uid string) {
	var in struct {
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	var hash string
	var thumb sql.NullString
	err := s.db.QueryRow(`SELECT password_hash, thumbnail FROM users WHERE id = ?`, uid).Scan(&hash, &thumb)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		writeError(w, http.StatusForbidden, "incorrect password")
		return
	}
	// Sessions and memberships cascade. Messages are kept with sender = NULL.
	if _, err := s.db.Exec(`DELETE FROM users WHERE id = ?`, uid); err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete account")
		return
	}
	if thumb.Valid {
		os.Remove(filepath.Join(s.dataDir, "thumbnails", thumb.String))
	}
	s.hub.kick(uid)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSetThumbnail(w http.ResponseWriter, r *http.Request, uid string) {
	f, err := s.receiveFile(w, r, maxThumbnailBytes, "thumbnails", true)
	if err != nil {
		writeUploadError(w, err, maxThumbnailBytes)
		return
	}
	var old sql.NullString
	s.db.QueryRow(`SELECT thumbnail FROM users WHERE id = ?`, uid).Scan(&old)

	if _, err := s.db.Exec(`UPDATE users SET thumbnail = ? WHERE id = ?`, f.diskName, uid); err != nil {
		os.Remove(f.fullPath)
		writeError(w, http.StatusInternalServerError, "could not save thumbnail")
		return
	}
	if old.Valid {
		os.Remove(filepath.Join(s.dataDir, "thumbnails", old.String))
	}
	user, _ := s.getUser(uid)
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleDeleteThumbnail(w http.ResponseWriter, r *http.Request, uid string) {
	var old sql.NullString
	s.db.QueryRow(`SELECT thumbnail FROM users WHERE id = ?`, uid).Scan(&old)
	if _, err := s.db.Exec(`UPDATE users SET thumbnail = NULL WHERE id = ?`, uid); err != nil {
		writeError(w, http.StatusInternalServerError, "update failed")
		return
	}
	if old.Valid {
		os.Remove(filepath.Join(s.dataDir, "thumbnails", old.String))
	}
	user, _ := s.getUser(uid)
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleGetThumbnail(w http.ResponseWriter, r *http.Request, _ string) {
	var name sql.NullString
	err := s.db.QueryRow(`SELECT thumbnail FROM users WHERE id = ?`, r.PathValue("id")).Scan(&name)
	if err != nil || !name.Valid {
		writeError(w, http.StatusNotFound, "no thumbnail")
		return
	}
	serveFile(w, r, filepath.Join(s.dataDir, "thumbnails", name.String),
		mime.TypeByExtension(filepath.Ext(name.String)))
}

// GET /api/users?q=sam  (prefix search on username or display name, for starting DMs)
func (s *Server) handleSearchUsers(w http.ResponseWriter, r *http.Request, uid string) {
	users := []*User{}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, users)
		return
	}
	like := escapeLike(q) + "%"
	rows, err := s.db.Query(
		`SELECT `+userCols+` FROM users u
		 WHERE u.id != ? AND (u.username LIKE ? ESCAPE '\' OR u.display_name LIKE ? ESCAPE '\')
		 ORDER BY u.username LIMIT 25`, uid, like, like)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}
	defer rows.Close()
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "search failed")
			return
		}
		users = append(users, u)
	}
	writeJSON(w, http.StatusOK, users)
}