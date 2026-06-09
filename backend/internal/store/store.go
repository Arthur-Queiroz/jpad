package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("note not found")

type Note struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	UpdatedAt int64  `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func Open(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS notes (
			path      TEXT PRIMARY KEY,
			content   TEXT NOT NULL DEFAULT '',
			updated_at INTEGER NOT NULL
		)
	`)
	return err
}

func (s *Store) Get(path string) (Note, error) {
	var n Note
	err := s.db.QueryRow(
		"SELECT path, content, updated_at FROM notes WHERE path = ?", path,
	).Scan(&n.Path, &n.Content, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, ErrNotFound
	}
	if err != nil {
		return Note{}, fmt.Errorf("get note: %w", err)
	}
	return n, nil
}

func (s *Store) Upsert(path, content string) (Note, error) {
	now := time.Now().Unix()
	_, err := s.db.Exec(`
		INSERT INTO notes (path, content, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET content = excluded.content, updated_at = excluded.updated_at
	`, path, content, now)
	if err != nil {
		return Note{}, fmt.Errorf("upsert note: %w", err)
	}
	return Note{Path: path, Content: content, UpdatedAt: now}, nil
}
