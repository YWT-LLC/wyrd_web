/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxContentRunes          = 4000
	maxAttachmentsPerMessage = 10
	maxGroupMembers          = 100
)

// clientError is safe to show to the user; anything else is logged and hidden.
type clientError string

func (e clientError) Error() string { return string(e) }

type Conversation struct {
	ID            string     `json:"id"`
	Type          string     `json:"type"` // bonfire | dm | group
	Name          string     `json:"name,omitempty"`
	Members       []*User    `json:"members,omitempty"` // empty for bonfires (everyone is in)
	CreatedAt     time.Time  `json:"created_at"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
}

type Message struct {
	ID             string       `json:"id"`
	ConversationID string       `json:"conversation_id"`
	Sender         *User        `json:"sender"` // null = deleted account
	Content        string       `json:"content"`
	Attachments    []Attachment `json:"attachments"`
	CreatedAt      time.Time    `json:"created_at"`
}

// ---- Access + membership helpers ----

func (s *Server) accessType(uid, convID string) (string, bool) {
	var t string
	err := s.db.QueryRow(`
		SELECT c.type FROM conversations c
		WHERE c.id = ? AND (c.type = 'bonfire' OR EXISTS (
			SELECT 1 FROM conversation_members m WHERE m.conversation_id = c.id AND m.user_id = ?))`,
		convID, uid).Scan(&t)
	return t, err == nil
}

func (s *Server) memberIDs(convID string) ([]string, error) {
	rows, err := s.db.Query(`SELECT user_id FROM conversation_members WHERE conversation_id = ?`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Server) loadMembers(convs []*Conversation) error {
	byID := map[string]*Conversation{}
	var args []any
	for _, c := range convs {
		if c.Type == "bonfire" {
			continue
		}
		byID[c.ID] = c
		args = append(args, c.ID)
	}
	if len(args) == 0 {
		return nil
	}
	rows, err := s.db.Query(
		`SELECT m.conversation_id, `+userCols+`
		 FROM conversation_members m JOIN users u ON u.id = m.user_id
		 WHERE m.conversation_id IN (`+placeholders(len(args))+`)
		 ORDER BY m.joined_at, u.username`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var convID string
		var u User
		var thumb sql.NullString
		var created int64
		if err := rows.Scan(&convID, &u.ID, &u.Username, &u.DisplayName, &thumb, &created); err != nil {
			return err
		}
		u.finish(thumb, created)
		byID[convID].Members = append(byID[convID].Members, &u)
	}
	return rows.Err()
}

func (s *Server) getConversation(id string) (*Conversation, error) {
	var c Conversation
	var created int64
	err := s.db.QueryRow(
		`SELECT id, type, COALESCE(name, ''), created_at FROM conversations WHERE id = ?`, id,
	).Scan(&c.ID, &c.Type, &c.Name, &created)
	if err != nil {
		return nil, err
	}
	c.CreatedAt = time.UnixMilli(created).UTC()
	if err := s.loadMembers([]*Conversation{&c}); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Server) userExists(id string) bool {
	var one int
	return s.db.QueryRow(`SELECT 1 FROM users WHERE id = ?`, id).Scan(&one) == nil
}

// ---- Conversation endpoints ----

// Returns the bonfire(s) plus every DM/group the user is in.
// The client sorts by last_message_at (falling back to created_at).
func (s *Server) handleListConversations(w http.ResponseWriter, r *http.Request, uid string) {
	rows, err := s.db.Query(`
		SELECT c.id, c.type, COALESCE(c.name, ''), c.created_at,
		       (SELECT MAX(created_at) FROM messages WHERE conversation_id = c.id)
		FROM conversations c
		WHERE c.type = 'bonfire' OR EXISTS (
			SELECT 1 FROM conversation_members m WHERE m.conversation_id = c.id AND m.user_id = ?)
		ORDER BY c.created_at`, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list conversations")
		return
	}
	defer rows.Close()

	convs := []*Conversation{}
	for rows.Next() {
		var c Conversation
		var created int64
		var last sql.NullInt64
		if err := rows.Scan(&c.ID, &c.Type, &c.Name, &created, &last); err != nil {
			writeError(w, http.StatusInternalServerError, "could not list conversations")
			return
		}
		c.CreatedAt = time.UnixMilli(created).UTC()
		if last.Valid {
			t := time.UnixMilli(last.Int64).UTC()
			c.LastMessageAt = &t
		}
		convs = append(convs, &c)
	}
	rows.Close()
	if err := s.loadMembers(convs); err != nil {
		writeError(w, http.StatusInternalServerError, "could not list conversations")
		return
	}
	writeJSON(w, http.StatusOK, convs)
}

// POST /api/conversations/dm {"user_id": "..."} - returns the existing DM if there is one.
func (s *Server) handleCreateDM(w http.ResponseWriter, r *http.Request, uid string) {
	var in struct {
		UserID string `json:"user_id"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.UserID == "" || in.UserID == uid {
		writeError(w, http.StatusBadRequest, "pick another user to message")
		return
	}
	if !s.userExists(in.UserID) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	a, b := uid, in.UserID
	if a > b {
		a, b = b, a
	}
	key := a + ":" + b

	tx, err := s.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create DM")
		return
	}
	defer tx.Rollback()

	id := newID()
	now := time.Now().UnixMilli()
	res, err := tx.Exec(
		`INSERT INTO conversations (id, type, dm_key, created_by, created_at)
		 VALUES (?, 'dm', ?, ?, ?) ON CONFLICT(dm_key) DO NOTHING`, id, key, uid, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create DM")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		tx.Rollback()
		var existing string
		if err := s.db.QueryRow(`SELECT id FROM conversations WHERE dm_key = ?`, key).Scan(&existing); err != nil {
			writeError(w, http.StatusInternalServerError, "could not load DM")
			return
		}
		conv, _ := s.getConversation(existing)
		writeJSON(w, http.StatusOK, conv)
		return
	}
	for _, m := range []string{uid, in.UserID} {
		if _, err := tx.Exec(
			`INSERT INTO conversation_members (conversation_id, user_id, joined_at) VALUES (?, ?, ?)`,
			id, m, now); err != nil {
			writeError(w, http.StatusInternalServerError, "could not create DM")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create DM")
		return
	}

	conv, _ := s.getConversation(id)
	s.hub.sendEvent([]string{in.UserID}, serverEvent{Type: "conversation.upsert", Conversation: conv})
	writeJSON(w, http.StatusCreated, conv)
}

// POST /api/conversations/group {"name": "...", "member_ids": ["..."]}
func (s *Server) handleCreateGroup(w http.ResponseWriter, r *http.Request, uid string) {
	var in struct {
		Name      string   `json:"name"`
		MemberIDs []string `json:"member_ids"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	name := strings.TrimSpace(in.Name)
	if n := utf8.RuneCountInString(name); n < 1 || n > 64 {
		writeError(w, http.StatusBadRequest, "group name must be 1-64 characters")
		return
	}

	set := map[string]bool{uid: true}
	for _, id := range in.MemberIDs {
		set[id] = true
	}
	if len(set) > maxGroupMembers {
		writeError(w, http.StatusBadRequest, "too many members")
		return
	}
	ids := make([]string, 0, len(set))
	args := make([]any, 0, len(set))
	for id := range set {
		ids = append(ids, id)
		args = append(args, id)
	}
	var found int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE id IN (`+placeholders(len(args))+`)`, args...,
	).Scan(&found)
	if err != nil || found != len(ids) {
		writeError(w, http.StatusBadRequest, "one or more members don't exist")
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create group")
		return
	}
	defer tx.Rollback()

	id := newID()
	now := time.Now().UnixMilli()
	if _, err := tx.Exec(
		`INSERT INTO conversations (id, type, name, created_by, created_at) VALUES (?, 'group', ?, ?, ?)`,
		id, name, uid, now); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create group")
		return
	}
	for _, m := range ids {
		if _, err := tx.Exec(
			`INSERT INTO conversation_members (conversation_id, user_id, joined_at) VALUES (?, ?, ?)`,
			id, m, now); err != nil {
			writeError(w, http.StatusInternalServerError, "could not create group")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create group")
		return
	}

	conv, _ := s.getConversation(id)
	s.hub.sendEvent(ids, serverEvent{Type: "conversation.upsert", Conversation: conv})
	writeJSON(w, http.StatusCreated, conv)
}

// POST /api/conversations/{id}/members {"user_id": "..."}  (groups only; any member can add)
func (s *Server) handleAddMember(w http.ResponseWriter, r *http.Request, uid string) {
	convID := r.PathValue("id")
	ctype, ok := s.accessType(uid, convID)
	if !ok {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if ctype != "group" {
		writeError(w, http.StatusBadRequest, "members can only be added to group chats")
		return
	}
	var in struct {
		UserID string `json:"user_id"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	if !s.userExists(in.UserID) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM conversation_members WHERE conversation_id = ?`, convID).Scan(&count)
	if count >= maxGroupMembers {
		writeError(w, http.StatusBadRequest, "group is full")
		return
	}
	if _, err := s.db.Exec(
		`INSERT OR IGNORE INTO conversation_members (conversation_id, user_id, joined_at) VALUES (?, ?, ?)`,
		convID, in.UserID, time.Now().UnixMilli()); err != nil {
		writeError(w, http.StatusInternalServerError, "could not add member")
		return
	}
	conv, _ := s.getConversation(convID)
	ids, _ := s.memberIDs(convID)
	s.hub.sendEvent(ids, serverEvent{Type: "conversation.upsert", Conversation: conv})
	writeJSON(w, http.StatusOK, conv)
}

// DELETE /api/conversations/{id}/members/me  (groups only)
func (s *Server) handleLeaveGroup(w http.ResponseWriter, r *http.Request, uid string) {
	convID := r.PathValue("id")
	ctype, ok := s.accessType(uid, convID)
	if !ok {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if ctype != "group" {
		writeError(w, http.StatusBadRequest, "you can only leave group chats")
		return
	}
	if _, err := s.db.Exec(
		`DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?`, convID, uid); err != nil {
		writeError(w, http.StatusInternalServerError, "could not leave group")
		return
	}
	if conv, err := s.getConversation(convID); err == nil {
		ids, _ := s.memberIDs(convID)
		s.hub.sendEvent(ids, serverEvent{Type: "conversation.upsert", Conversation: conv})
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- Messages ----

const messageSelect = `
	SELECT m.id, m.conversation_id, m.content, m.created_at,
	       u.id, u.username, u.display_name, u.thumbnail, u.created_at
	FROM messages m LEFT JOIN users u ON u.id = m.sender_id`

func (s *Server) queryMessages(query string, args ...any) ([]*Message, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []*Message
	for rows.Next() {
		var m Message
		var created int64
		var uid, uname, dname, thumb sql.NullString
		var ucreated sql.NullInt64
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Content, &created,
			&uid, &uname, &dname, &thumb, &ucreated); err != nil {
			return nil, err
		}
		m.CreatedAt = time.UnixMilli(created).UTC()
		if uid.Valid {
			u := &User{ID: uid.String, Username: uname.String, DisplayName: dname.String}
			u.finish(thumb, ucreated.Int64)
			m.Sender = u
		}
		m.Attachments = []Attachment{}
		msgs = append(msgs, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	return msgs, s.loadAttachments(msgs)
}

func (s *Server) loadAttachments(msgs []*Message) error {
	if len(msgs) == 0 {
		return nil
	}
	byID := make(map[string]*Message, len(msgs))
	args := make([]any, 0, len(msgs))
	for _, m := range msgs {
		byID[m.ID] = m
		args = append(args, m.ID)
	}
	rows, err := s.db.Query(
		`SELECT message_id, id, kind, mime, size, filename FROM attachments
		 WHERE message_id IN (`+placeholders(len(args))+`) ORDER BY created_at, id`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var mid string
		var a Attachment
		if err := rows.Scan(&mid, &a.ID, &a.Kind, &a.Mime, &a.Size, &a.Filename); err != nil {
			return err
		}
		a.URL = "/api/files/" + a.ID
		byID[mid].Attachments = append(byID[mid].Attachments, a)
	}
	return rows.Err()
}

// GET /api/conversations/{id}/messages?before=<message id>&limit=50
// Returns one page, oldest -> newest. Pass the first message's id as `before` to go further back.
func (s *Server) handleListMessages(w http.ResponseWriter, r *http.Request, uid string) {
	convID := r.PathValue("id")
	if _, ok := s.accessType(uid, convID); !ok {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	query := messageSelect + ` WHERE m.conversation_id = ?`
	args := []any{convID}
	if before := r.URL.Query().Get("before"); before != "" {
		query += ` AND m.seq < (SELECT seq FROM messages WHERE id = ?)`
		args = append(args, before)
	}
	query += ` ORDER BY m.seq DESC LIMIT ?`
	args = append(args, limit)

	msgs, err := s.queryMessages(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load messages")
		return
	}
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	if msgs == nil {
		msgs = []*Message{}
	}
	writeJSON(w, http.StatusOK, msgs)
}

// postMessage validates, persists, and fans out a message. Used by the WebSocket handler.
func (s *Server) postMessage(senderID, convID, content string, attachmentIDs []string) error {
	content = strings.TrimSpace(content)
	if content == "" && len(attachmentIDs) == 0 {
		return clientError("message is empty")
	}
	if utf8.RuneCountInString(content) > maxContentRunes {
		return clientError("message is too long")
	}
	if len(attachmentIDs) > maxAttachmentsPerMessage {
		return clientError("too many attachments")
	}
	convType, ok := s.accessType(senderID, convID)
	if !ok {
		return clientError("conversation not found")
	}

	id := s.baseURL + "/notes/" + newID() // URI-style ID, ActivityPub-friendly like before
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`INSERT INTO messages (id, conversation_id, sender_id, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, convID, senderID, content, time.Now().UnixMilli()); err != nil {
		return err
	}
	for _, aid := range attachmentIDs {
		res, err := tx.Exec(
			`UPDATE attachments SET message_id = ? WHERE id = ? AND uploader_id = ? AND message_id IS NULL`,
			id, aid, senderID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n != 1 {
			return clientError("invalid or already-used attachment")
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	msgs, err := s.queryMessages(messageSelect+` WHERE m.id = ?`, id)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		return errors.New("message not found after insert")
	}

	ev := serverEvent{Type: "message.new", Message: msgs[0]}
	if convType == "bonfire" {
		s.hub.broadcast(ev)
		return nil
	}
	ids, err := s.memberIDs(convID)
	if err != nil {
		return err
	}
	s.hub.sendEvent(ids, ev)
	return nil
}