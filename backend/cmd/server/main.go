package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nishoof/flexi/backend/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth", handler.AuthHandler)
	mux.HandleFunc("/api/entries", handler.EntriesHandler)
	mux.HandleFunc("/api/terms", handler.TermsHandler)
	mux.HandleFunc("/api/terms/", handler.TermsHandler)

	log.Printf("starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
