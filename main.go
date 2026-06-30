package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Harshupadhyay-10/uptimer/internal/api"
	"github.com/Harshupadhyay-10/uptimer/internal/scheduler"
	"github.com/Harshupadhyay-10/uptimer/internal/storage"
)

func main() {
	store, err := storage.New("uptimer.db")
	if err != nil {
		log.Fatal("failed to open storage:", err)
	}
	defer store.Close()

	// Start background scheduler — checks every 30 seconds
	sched := scheduler.New(store, 30*time.Second)
	sched.Start()
	defer sched.Stop()

	// Start REST API
	mux := http.NewServeMux()
	server := api.New(store)
	server.RegisterRoutes(mux)

	// Graceful shutdown on Ctrl+C
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	log.Println("Uptimer API running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
