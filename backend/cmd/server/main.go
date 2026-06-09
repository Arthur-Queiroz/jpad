package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Arthur-Queiroz/jpad/internal/api"
	"github.com/Arthur-Queiroz/jpad/internal/store"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "notes.db"
	}

	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	h := api.NewHandler(s)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("servidor iniciado na porta %s\n", port)
	if err := http.ListenAndServe(":"+port, h.Routes()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
