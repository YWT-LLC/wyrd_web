/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionTTL = 30 * 24 * time.Hour

var (
	usernameRe = regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`)
	// Compared against when a username doesn't exist, so response time doesn't reveal valid usernames.
	dummyHash, _ = bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)
)

// ---- JSON helpers ----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

// ---- Sessions ----

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

func (s *Server) createSession(userID string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	now := time.Now()
	_, err := s.db.Exec(
		`INSERT INTO sessions (token_hash, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`,
		hashToken(token), userID, now.UnixMilli(), now.Add(sessionTTL).UnixMilli(),
	)
	return token, err
}

func bearerToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

func (s *Server) userForToken(token string) (string, bool) {
	if token == "" {
		return "", false
	}
	var uid string
	err := s.db.QueryRow(
		`SELECT user_id FROM sessions WHERE token_hash = ? AND expires_at > ?`,
		hashToken(token), time.Now().UnixMilli(),
	).Scan(&uid)
	return uid, err == nil
}

type authedHandler func(w http.ResponseWriter, r *http.Request, uid string)

func (s *Server) auth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := s.userForToken(bearerToken(r))
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r, uid)
	}
}

// ---- Handlers ----

type credentials struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	InviteCode  string `json:"invite_code"`
}

type authResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	if s.inviteCode != "" &&
		subtle.ConstantTimeCompare([]byte(in.InviteCode), []byte(s.inviteCode)) != 1 {
		writeError(w, http.StatusForbidden, "invalid invite code")
		return
	}

	in.Username = strings.TrimSpace(in.Username)
	if !usernameRe.MatchString(in.Username) {
		writeError(w, http.StatusBadRequest, "username must be 3-32 characters: letters, numbers, underscore")
		return
	}
	// bcrypt only uses the first 72 bytes, so cap there.
	if len(in.Password) < 8 || len(in.Password) > 72 {
		writeError(w, http.StatusBadRequest, "password must be 8-72 bytes")
		return
	}
	display := in.DisplayName
	if strings.TrimSpace(display) == "" {
		display = in.Username
	}
	display, ok := validDisplayName(display)
	if !ok {
		writeError(w, http.StatusBadRequest, "display name must be 1-64 characters")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create account")
		return
	}
	id := newID()
	_, err = s.db.Exec(
		`INSERT INTO users (id, username, password_hash, display_name, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, in.Username, string(hash), display, time.Now().UnixMilli(),
	)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "username is taken")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create account")
		return
	}

	token, err := s.createSession(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	user, _ := s.getUser(id)
	writeJSON(w, http.StatusCreated, authResponse{Token: token, User: user})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	var id, hash string
	err := s.db.QueryRow(
		`SELECT id, password_hash FROM users WHERE username = ?`, strings.TrimSpace(in.Username),
	).Scan(&id, &hash)
	if err != nil {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(in.Password))
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	token, err := s.createSession(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	user, _ := s.getUser(id)
	writeJSON(w, http.StatusOK, authResponse{Token: token, User: user})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request, uid string) {
	s.db.Exec(`DELETE FROM sessions WHERE token_hash = ?`, hashToken(bearerToken(r)))
	w.WriteHeader(http.StatusNoContent)
}