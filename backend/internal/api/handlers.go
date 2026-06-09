package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/Arthur-Queiroz/jpad/internal/store"
	"github.com/go-chi/chi/v5"
)

const maxBodySize = 1 << 20 // 1MB

var pathRe = regexp.MustCompile(`^[a-zA-Z0-9_/-]+$`)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/api/*", h.getNote)
	r.Put("/api/*", h.putNote)
	return r
}

func (h *Handler) getNote(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "*")
	log.Printf("GET path=%q", path)
	if err := validatePath(path); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	n, err := h.store.Get(path)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusOK, store.Note{Path: path, Content: "", UpdatedAt: 0})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) putNote(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "*")
	if err := validatePath(path); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) || err.Error() == "http: request body too large" {
			writeError(w, http.StatusRequestEntityTooLarge, "body exceeds 1MB limit")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	n, err := h.store.Upsert(path, body.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, n)
}

func validatePath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("path is required")
	}
	if !pathRe.MatchString(path) {
		return errors.New("path contains invalid characters")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
