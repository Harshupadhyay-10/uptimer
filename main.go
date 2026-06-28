package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Harshupadhyay-10/uptimer/internal/checker"
	"github.com/Harshupadhyay-10/uptimer/internal/storage"
)

func main() {
	store, err := storage.New("uptimer.db")
	if err != nil {
		log.Fatal("failed to open storage:", err)
	}
	defer store.Close()

	urls := []string{
		"https://example.com",
		"https://google.com",
		"https://github.com",
		"https://this-domain-definitely-does-not-exist-12345.com",
	}

	results := checker.CheckAll(urls, 5*time.Second)

	for _, r := range results {
		if err := store.SaveCheck(r.URL, r.StatusCode, r.Latency, r.Err, r.CheckedAt); err != nil {
			log.Println("failed to save check:", err)
		}

		if r.Err != nil {
			fmt.Printf("URL: %s | ERROR: %v\n", r.URL, r.Err)
			continue
		}
		fmt.Printf("URL: %s | Status: %d | Latency: %v\n", r.URL, r.StatusCode, r.Latency)
	}
}
