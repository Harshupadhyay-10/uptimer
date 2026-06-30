package main

import (
	"log"
	"net/http"

	"github.com/Harshupadhyay-10/uptimer/internal/api"
	"github.com/Harshupadhyay-10/uptimer/internal/storage"
)

func main() {
	store, err := storage.New("uptimer.db")
	if err != nil {
		log.Fatal("failed to open storage:", err)
	}
	defer store.Close()

	mux := http.NewServeMux()
	server := api.New(store)
	server.RegisterRoutes(mux)

	log.Println("Uptimer API running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
