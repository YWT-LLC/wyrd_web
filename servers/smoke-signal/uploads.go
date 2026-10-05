/* wyrd_web
 * Copyright (c) 2026 YWT (Empathetech LLC). All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Images are verified by sniffing the actual bytes. Audio/video containers are
// hard to sniff reliably (m4a vs mp4, ogg, ...), so for those we accept the
// client-declared type if it's on the allowlist and the bytes aren't text.
var imageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var avTypes = map[string]struct{ kind, ext string }{
	"video/mp4":       {"video", ".mp4"},
	"video/webm":      {"video", ".webm"},
	"video/quicktime": {"video", ".mov"},
	"audio/mpeg":      {"audio", ".mp3"},
	"audio/mp4":       {"audio", ".m4a"},
	"audio/aac":       {"audio", ".aac"},
	"audio/ogg":       {"audio", ".ogg"},
	"audio/wav":       {"audio", ".wav"},
	"audio/webm":      {"audio", ".weba"},
}

var mimeAliases = map[string]string{
	"audio/wave":  "audio/wav",
	"audio/x-wav": "audio/wav",
	"audio/x-m4a": "audio/mp4",
	"audio/mp3":   "audio/mpeg",
}

var (
	errTooLarge   = errors.New("file too large")
	errBadType    = errors.New("unsupported file type")
	errNoFile     = errors.New(`expected a multipart field named "file"`)
	errEmptyFile  = errors.New("file is empty")
	errBadRequest = errors.New("malformed upload")
)

type Attachment struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"` // image | audio | video
	Mime     string `json:"mime"`
	Size     int64  `json:"size"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

type storedFile struct {
	id, diskName, fullPath, mime, kind, origName string
	size                                         int64
}

func baseMIME(v string) string {
	v = strings.ToLower(strings.TrimSpace(strings.SplitN(v, ";", 2)[0]))
	if alias, ok := mimeAliases[v]; ok {
		return alias
	}
	return v
}

func classify(sniffed, declared string, imagesOnly bool) (mimeType, kind, ext string, ok bool) {
	sniffed, declared = baseMIME(sniffed), baseMIME(declared)
	if e, found := imageTypes[sniffed]; found {
		return sniffed, "image", e, true
	}
	if imagesOnly || strings.HasPrefix(sniffed, "text/") {
		return "", "", "", false
	}
	if t, found := avTypes[declared]; found {
		return declared, t.kind, t.ext, true
	}
	if t, found := avTypes[sniffed]; found {
		return sniffed, t.kind, t.ext, true
	}
	return "", "", "", false
}

func bodyErr(err error) error {
	var mbe *http.MaxBytesError
	if errors.As(err, &mbe) {
		return errTooLarge
	}
	return err
}

// receiveFile streams the multipart field "file" to disk without buffering it in memory.
func (s *Server) receiveFile(w http.ResponseWriter, r *http.Request, limit int64, subdir string, imagesOnly bool) (*storedFile, error) {
	// Slack for multipart framing; the exact file limit is enforced below.
	r.Body = http.MaxBytesReader(w, r.Body, limit+(1<<20))
	mr, err := r.MultipartReader()
	if err != nil {
		return nil, errBadRequest
	}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			return nil, errNoFile
		}
		if err != nil {
			if bodyErr(err) == errTooLarge {
				return nil, errTooLarge
			}
			return nil, errBadRequest
		}
		if part.FormName() == "file" {
			return s.storePart(part, limit, subdir, imagesOnly)
		}
	}
}

func (s *Server) storePart(part *multipart.Part, limit int64, subdir string, imagesOnly bool) (*storedFile, error) {
	br := bufio.NewReaderSize(part, 512)
	head, err := br.Peek(512)
	if err != nil && err != io.EOF {
		return nil, bodyErr(err)
	}
	if len(head) == 0 {
		return nil, errEmptyFile
	}

	mimeType, kind, ext, ok := classify(http.DetectContentType(head), part.Header.Get("Content-Type"), imagesOnly)
	if !ok {
		return nil, errBadType
	}

	id := newID()
	diskName := id + ext
	fullPath := filepath.Join(s.dataDir, subdir, diskName)
	out, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, err
	}
	n, err := io.Copy(out, io.LimitReader(br, limit+1))
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(fullPath)
		return nil, bodyErr(err)
	}
	if n > limit {
		os.Remove(fullPath)
		return nil, errTooLarge
	}

	name := filepath.Base(part.FileName())
	if name == "." || name == "/" || name == "" {
		name = "file" + ext
	}
	if rs := []rune(name); len(rs) > 100 {
		name = string(rs[:100])
	}
	return &storedFile{
		id: id, diskName: diskName, fullPath: fullPath,
		mime: mimeType, kind: kind, origName: name, size: n,
	}, nil
}

func writeUploadError(w http.ResponseWriter, err error, limit int64) {
	switch {
	case errors.Is(err, errTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("file exceeds the %d MB limit", limit>>20))
	case errors.Is(err, errBadType):
		writeError(w, http.StatusUnsupportedMediaType, err.Error())
	case errors.Is(err, errNoFile), errors.Is(err, errEmptyFile), errors.Is(err, errBadRequest):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("upload error: %v", err)
		writeError(w, http.StatusInternalServerError, "upload failed")
	}
}

// POST /api/uploads (multipart, field "file") -> Attachment
// The attachment is "pending" until a message references it.
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request, uid string) {
	f, err := s.receiveFile(w, r, maxMediaBytes, "uploads", false)
	if err != nil {
		writeUploadError(w, err, maxMediaBytes)
		return
	}
	_, err = s.db.Exec(
		`INSERT INTO attachments (id, uploader_id, kind, mime, size, filename, path, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		f.id, uid, f.kind, f.mime, f.size, f.origName, f.diskName, time.Now().UnixMilli(),
	)
	if err != nil {
		os.Remove(f.fullPath)
		writeError(w, http.StatusInternalServerError, "could not save upload")
		return
	}
	writeJSON(w, http.StatusCreated, Attachment{
		ID: f.id, Kind: f.kind, Mime: f.mime, Size: f.size,
		Filename: f.origName, URL: "/api/files/" + f.id,
	})
}

// GET /api/files/{id} - only the uploader (while pending) or members of the
// message's conversation (anyone, for bonfires) can fetch it.
func (s *Server) handleGetFile(w http.ResponseWriter, r *http.Request, uid string) {
	var diskName, mimeType string
	err := s.db.QueryRow(`
		SELECT a.path, a.mime
		FROM attachments a
		LEFT JOIN messages m ON m.id = a.message_id
		LEFT JOIN conversations c ON c.id = m.conversation_id
		WHERE a.id = ? AND (
			(a.message_id IS NULL AND a.uploader_id = ?)
			OR c.type = 'bonfire'
			OR EXISTS (SELECT 1 FROM conversation_members cm
			           WHERE cm.conversation_id = c.id AND cm.user_id = ?)
		)`, r.PathValue("id"), uid, uid).Scan(&diskName, &mimeType)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	serveFile(w, r, filepath.Join(s.dataDir, "uploads", diskName), mimeType)
}

// serveFile uses ServeContent so Range requests work (needed for video seeking).
func serveFile(w http.ResponseWriter, r *http.Request, fullPath, contentType string) {
	f, err := os.Open(fullPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "file error")
		return
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	http.ServeContent(w, r, "", info.ModTime(), f)
}
