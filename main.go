package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Harshupadhyay-10/uptimer/internal/checker"
	"github.com/Harshupadhyay-10/uptimer/internal/storage"
)

func main() {
	// Delete the old uptimer.db first since schema changed
	store, err := storage.New("uptimer.db")
	if err != nil {
		log.Fatal("failed to open storage:", err)
	}
	defer store.Close()

	// Add some targets to monitor
	urls := []string{
		"https://example.com",
		"https://google.com",
		"https://github.com",
		"https://this-domain-definitely-does-not-exist-12345.com",
	}

	// Insert targets into DB (ignore error if already exists)
	targetIDs := make(map[string]int64)
	for _, url := range urls {
		id, err := store.AddTarget(url)
		if err != nil {
			log.Printf("skipping %s (may already exist): %v", url, err)
			continue
		}
		targetIDs[url] = id
	}

	// Run concurrent checks
	results := checker.CheckAll(urls, 5*time.Second)

	for _, r := range results {
		id, ok := targetIDs[r.URL]
		if !ok {
			log.Printf("no target ID found for %s, skipping save", r.URL)
			continue
		}

		if err := store.SaveCheck(id, r.StatusCode, r.Latency, r.Err, r.CheckedAt); err != nil {
			log.Println("failed to save check:", err)
		}

		if r.Err != nil {
			fmt.Printf("URL: %s | ERROR: %v\n", r.URL, r.Err)
			continue
		}
		fmt.Printf("URL: %s | Status: %d | Latency: %v\n", r.URL, r.StatusCode, r.Latency)
	}
}
