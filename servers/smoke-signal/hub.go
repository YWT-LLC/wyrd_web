/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Server -> client events
type serverEvent struct {
	Type         string        `json:"type"` // message.new | conversation.upsert | error
	Message      *Message      `json:"message,omitempty"`
	Conversation *Conversation `json:"conversation,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// Client -> server events
type clientEvent struct {
	Type           string   `json:"type"` // message.send
	ConversationID string   `json:"conversation_id"`
	Content        string   `json:"content"`
	AttachmentIDs  []string `json:"attachment_ids"`
}

type Client struct {
	userID string
	conn   *websocket.Conn
	send   chan []byte
	cancel context.CancelFunc
}

// enqueue never blocks. A client that can't keep up is dropped (it reconnects and refetches).
func (c *Client) enqueue(data []byte) {
	select {
	case c.send <- data:
	default:
		c.cancel()
	}
}

func (c *Client) sendError(msg string) {
	data, _ := json.Marshal(serverEvent{Type: "error", Error: msg})
	c.enqueue(data)
}

func (c *Client) writePump(ctx context.Context) {
	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-c.send:
			wctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.conn.Write(wctx, websocket.MessageText, data)
			cancel()
			if err != nil {
				c.cancel()
				return
			}
		case <-ping.C:
			pctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.conn.Ping(pctx)
			cancel()
			if err != nil {
				c.cancel()
				return
			}
		}
	}
}

// Hub tracks live connections per user (a user can be connected from several devices).
type Hub struct {
	mu    sync.RWMutex
	conns map[string]map[*Client]struct{}
}

func newHub() *Hub {
	return &Hub{conns: make(map[string]map[*Client]struct{})}
}

func (h *Hub) add(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[c.userID] == nil {
		h.conns[c.userID] = make(map[*Client]struct{})
	}
	h.conns[c.userID][c] = struct{}{}
}

func (h *Hub) remove(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns[c.userID], c)
	if len(h.conns[c.userID]) == 0 {
		delete(h.conns, c.userID)
	}
}

func (h *Hub) sendEvent(userIDs []string, ev serverEvent) {
	data, err := json.Marshal(ev)
	if err != nil {
		log.Printf("marshal event: %v", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, uid := range userIDs {
		for c := range h.conns[uid] {
			c.enqueue(data)
		}
	}
}

func (h *Hub) broadcast(ev serverEvent) {
	data, err := json.Marshal(ev)
	if err != nil {
		log.Printf("marshal event: %v", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, set := range h.conns {
		for c := range set {
			c.enqueue(data)
		}
	}
}

// kick closes every connection a user has (used on account deletion).
func (h *Hub) kick(userID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.conns[userID] {
		c.cancel()
	}
}

// GET /ws?token=...  (or an Authorization: Bearer header)
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if token == "" {
		token = r.URL.Query().Get("token") // browsers/Flutter web can't set WS headers
	}
	uid, ok := s.userForToken(token)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println("websocket accept:", err)
		return
	}
	conn.SetReadLimit(64 << 10) // messages are small; media goes through /api/uploads

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	c := &Client{userID: uid, conn: conn, send: make(chan []byte, 64), cancel: cancel}
	s.hub.add(c)
	defer s.hub.remove(c)
	go c.writePump(ctx)

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			break
		}
		s.handleClientEvent(c, data)
	}
	conn.Close(websocket.StatusNormalClosure, "")
}

func (s *Server) handleClientEvent(c *Client, data []byte) {
	var ev clientEvent
	if err := json.Unmarshal(data, &ev); err != nil {
		c.sendError("invalid event")
		return
	}
	switch ev.Type {
	case "message.send":
		if err := s.postMessage(c.userID, ev.ConversationID, ev.Content, ev.AttachmentIDs); err != nil {
			var ce clientError
			if errors.As(err, &ce) {
				c.sendError(string(ce))
			} else {
				log.Printf("post message: %v", err)
				c.sendError("internal error")
			}
		}
	default:
		c.sendError("unknown event type")
	}
}