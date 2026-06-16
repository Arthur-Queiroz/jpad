package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Arthur-Queiroz/jpad/internal/store"
)

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return NewHandler(s)
}

func doRequest(t *testing.T, h *Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)
	return rr
}

type noteResponse struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	UpdatedAt int64  `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func decodeNote(t *testing.T, rr *httptest.ResponseRecorder) noteResponse {
	t.Helper()
	var n noteResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &n); err != nil {
		t.Fatalf("failed to decode response: %v\nbody: %s", err, rr.Body.String())
	}
	return n
}

func decodeError(t *testing.T, rr *httptest.ResponseRecorder) errorResponse {
	t.Helper()
	var e errorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &e); err != nil {
		t.Fatalf("failed to decode error response: %v\nbody: %s", err, rr.Body.String())
	}
	return e
}

// --- GET tests ---

func TestGetNoteNotFound(t *testing.T) {
	h := newTestHandler(t)

	rr := doRequest(t, h, http.MethodGet, "/api/new-note", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	n := decodeNote(t, rr)
	if n.Content != "" {
		t.Errorf("Content = %q, want empty", n.Content)
	}
	if n.Path != "new-note" {
		t.Errorf("Path = %q, want %q", n.Path, "new-note")
	}
	if n.UpdatedAt != 0 {
		t.Errorf("UpdatedAt = %d, want 0", n.UpdatedAt)
	}
}

func TestGetNoteAfterPut(t *testing.T) {
	h := newTestHandler(t)

	putBody := map[string]string{"content": "hello world"}
	rr := doRequest(t, h, http.MethodPut, "/api/my-note", putBody)
	if rr.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want %d", rr.Code, http.StatusOK)
	}

	rr = doRequest(t, h, http.MethodGet, "/api/my-note", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", rr.Code, http.StatusOK)
	}

	n := decodeNote(t, rr)
	if n.Content != "hello world" {
		t.Errorf("Content = %q, want %q", n.Content, "hello world")
	}
	if n.Path != "my-note" {
		t.Errorf("Path = %q, want %q", n.Path, "my-note")
	}
}

func TestGetNoteReturnsLatestContent(t *testing.T) {
	h := newTestHandler(t)

	updates := []string{"first", "second", "third"}
	for _, c := range updates {
		doRequest(t, h, http.MethodPut, "/api/note", map[string]string{"content": c})
	}

	rr := doRequest(t, h, http.MethodGet, "/api/note", nil)
	n := decodeNote(t, rr)
	if n.Content != "third" {
		t.Errorf("Content = %q, want %q", n.Content, "third")
	}
}

// --- PUT tests ---

func TestPutNoteCreatesNew(t *testing.T) {
	h := newTestHandler(t)

	body := map[string]string{"content": "new note content"}
	rr := doRequest(t, h, http.MethodPut, "/api/new-note", body)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	n := decodeNote(t, rr)
	if n.Path != "new-note" {
		t.Errorf("Path = %q, want %q", n.Path, "new-note")
	}
	if n.Content != "new note content" {
		t.Errorf("Content = %q, want %q", n.Content, "new note content")
	}
	if n.UpdatedAt == 0 {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestPutNoteUpdatesExisting(t *testing.T) {
	h := newTestHandler(t)

	doRequest(t, h, http.MethodPut, "/api/note", map[string]string{"content": "v1"})
	rr := doRequest(t, h, http.MethodPut, "/api/note", map[string]string{"content": "v2"})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	n := decodeNote(t, rr)
	if n.Content != "v2" {
		t.Errorf("Content = %q, want %q", n.Content, "v2")
	}
}

func TestPutNoteEmptyContent(t *testing.T) {
	h := newTestHandler(t)

	rr := doRequest(t, h, http.MethodPut, "/api/note", map[string]string{"content": ""})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	n := decodeNote(t, rr)
	if n.Content != "" {
		t.Errorf("Content = %q, want empty", n.Content)
	}
}

func TestPutNoteInvalidJSON(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPut, "/api/note", strings.NewReader(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	e := decodeError(t, rr)
	if e.Error != "invalid JSON body" {
		t.Errorf("error = %q, want %q", e.Error, "invalid JSON body")
	}
}

func TestPutNoteBodyTooLarge(t *testing.T) {
	h := newTestHandler(t)

	// 1MB limit = 1<<20 bytes; JSON overhead: {"content":"..."}
	largeContent := strings.Repeat("x", (1<<20)+1)
	body, _ := json.Marshal(map[string]string{"content": largeContent})
	req := httptest.NewRequest(http.MethodPut, "/api/note", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusRequestEntityTooLarge)
	}

	e := decodeError(t, rr)
	if !strings.Contains(e.Error, "1MB") {
		t.Errorf("error = %q, want to contain '1MB'", e.Error)
	}
}

func TestPutNoteBodyJustBelowLimit(t *testing.T) {
	h := newTestHandler(t)

	// JSON wrapper is ~14 bytes, so use 1MB - 100 bytes to stay safely under
	content := strings.Repeat("x", (1<<20)-100)
	body, _ := json.Marshal(map[string]string{"content": content})
	req := httptest.NewRequest(http.MethodPut, "/api/note", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body size: %d bytes)", rr.Code, http.StatusOK, len(body))
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newTestHandler(t)

	methods := []string{http.MethodDelete, http.MethodPost, http.MethodPatch}
	for _, m := range methods {
		rr := doRequest(t, h, m, "/api/note", nil)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /api/note: status = %d, want %d", m, rr.Code, http.StatusMethodNotAllowed)
		}
	}
}

func TestPutNoteEmptyBody(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPut, "/api/note", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// --- Path validation tests ---

func TestGetInvalidPath(t *testing.T) {
	h := newTestHandler(t)

	rr := doRequest(t, h, http.MethodGet, "/api/bad@path", nil)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	e := decodeError(t, rr)
	if !strings.Contains(e.Error, "invalid characters") {
		t.Errorf("error = %q, want to contain 'invalid characters'", e.Error)
	}
}

func TestPutInvalidPath(t *testing.T) {
	h := newTestHandler(t)

	rr := doRequest(t, h, http.MethodPut, "/api/bad%20path", map[string]string{"content": "x"})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetPathWithSpecialChars(t *testing.T) {
	h := newTestHandler(t)

	badPaths := []string{"/api/a@b", "/api/a%20b", "/api/a!b", "/api/a.b"}
	for _, p := range badPaths {
		rr := doRequest(t, h, http.MethodGet, p, nil)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("GET %s: status = %d, want %d", p, rr.Code, http.StatusBadRequest)
		}
	}
}

func TestGetValidPaths(t *testing.T) {
	h := newTestHandler(t)

	validPaths := []string{"/api/hello", "/api/hello-world", "/api/hello_world", "/api/a/b/c", "/api/HelLo123"}
	for _, p := range validPaths {
		rr := doRequest(t, h, http.MethodGet, p, nil)
		if rr.Code != http.StatusOK {
			t.Errorf("GET %s: status = %d, want %d", p, rr.Code, http.StatusOK)
		}
	}
}

// --- Content-Type tests ---

func TestResponseContentTypeIsJSON(t *testing.T) {
	h := newTestHandler(t)

	rr := doRequest(t, h, http.MethodGet, "/api/note", nil)
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

// --- validatePath unit tests ---

func TestValidatePath(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		{"hello", false},
		{"hello-world", false},
		{"hello_world", false},
		{"a/b/c", false},
		{"A123", false},
		{"", true},
		{"bad@path", true},
		{"bad path", true},
		{"bad!path", true},
		{"bad.path", true},
	}

	for _, tt := range tests {
		err := validatePath(tt.path)
		if (err != nil) != tt.wantErr {
			t.Errorf("validatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
		}
	}
}
