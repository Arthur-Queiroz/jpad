package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestOpenAndClose(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}

	// Verify db file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
}

func TestOpenInvalidPath(t *testing.T) {
	_, err := Open("/nonexistent/dir/test.db")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestGetNotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.Get("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestUpsertCreatesNewNote(t *testing.T) {
	s := newTestStore(t)

	n, err := s.Upsert("my-note", "hello world")
	if err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	if n.Path != "my-note" {
		t.Errorf("Path = %q, want %q", n.Path, "my-note")
	}
	if n.Content != "hello world" {
		t.Errorf("Content = %q, want %q", n.Content, "hello world")
	}
	if n.UpdatedAt == 0 {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestUpsertUpdatesExistingNote(t *testing.T) {
	s := newTestStore(t)

	first, err := s.Upsert("my-note", "first version")
	if err != nil {
		t.Fatalf("first Upsert() error: %v", err)
	}

	second, err := s.Upsert("my-note", "second version")
	if err != nil {
		t.Fatalf("second Upsert() error: %v", err)
	}

	if second.Content != "second version" {
		t.Errorf("Content = %q, want %q", second.Content, "second version")
	}
	if second.Path != "my-note" {
		t.Errorf("Path = %q, want %q", second.Path, "my-note")
	}
	if second.UpdatedAt == first.UpdatedAt {
		// Technically possible if sub-second, but unlikely
		t.Log("warning: UpdatedAt did not change between upserts")
	}
}

func TestGetAfterUpsert(t *testing.T) {
	s := newTestStore(t)

	_, err := s.Upsert("test-path", "test content")
	if err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	n, err := s.Get("test-path")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if n.Path != "test-path" {
		t.Errorf("Path = %q, want %q", n.Path, "test-path")
	}
	if n.Content != "test content" {
		t.Errorf("Content = %q, want %q", n.Content, "test content")
	}
	if n.UpdatedAt == 0 {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestGetReturnsLatestAfterMultipleUpserts(t *testing.T) {
	s := newTestStore(t)

	contents := []string{"first", "second", "third"}
	for _, c := range contents {
		if _, err := s.Upsert("note", c); err != nil {
			t.Fatalf("Upsert(%q) error: %v", c, err)
		}
	}

	n, err := s.Get("note")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if n.Content != "third" {
		t.Errorf("Content = %q, want %q", n.Content, "third")
	}
}

func TestMultipleIndependentPaths(t *testing.T) {
	s := newTestStore(t)

	paths := []string{"notes/a", "notes/b", "other"}
	for i, p := range paths {
		content := strings.Repeat("x", i+1)
		if _, err := s.Upsert(p, content); err != nil {
			t.Fatalf("Upsert(%q) error: %v", p, err)
		}
	}

	for i, p := range paths {
		want := strings.Repeat("x", i+1)
		n, err := s.Get(p)
		if err != nil {
			t.Fatalf("Get(%q) error: %v", p, err)
		}
		if n.Content != want {
			t.Errorf("Get(%q).Content = %q, want %q", p, n.Content, want)
		}
	}
}

func TestUpsertEmptyContent(t *testing.T) {
	s := newTestStore(t)

	n, err := s.Upsert("empty", "")
	if err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}
	if n.Content != "" {
		t.Errorf("Content = %q, want empty", n.Content)
	}

	got, err := s.Get("empty")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got.Content != "" {
		t.Errorf("Get().Content = %q, want empty", got.Content)
	}
}

func TestUpsertOverwriteWithEmptyContent(t *testing.T) {
	s := newTestStore(t)

	s.Upsert("note", "some content")

	_, err := s.Upsert("note", "")
	if err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	n, err := s.Get("note")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if n.Content != "" {
		t.Errorf("Content = %q, want empty after overwrite", n.Content)
	}
}
