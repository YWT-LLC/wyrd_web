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

func (s *Server) storePart(part io.Reader, limit int64, subdir string, imagesOnly bool) (*storedFile, error) {
	type namedPart interface {
		FileName() string
		Header() map[string][]string
	}
	// multipart.Part exposes Header as a field, so we take the pieces we need via a wrapper in receiveFile's caller.
	_ = namedPart(nil)
	return nil, errors.New("replaced below")
}