package main

import (
	"fmt"
	"time"

	"github.com/Harshupadhyay-10/uptimer/internal/checker"
)

func main() {
	urls := []string{
		"https://example.com",
		"https://google.com",
		"https://github.com",
		"https://this-domain-definitely-does-not-exist-12345.com",
	}

	results := checker.CheckAll(urls, 5*time.Second)

	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("URL: %s | ERROR: %v\n", r.URL, r.Err)
			continue
		}
		fmt.Printf("URL: %s | Status: %d | Latency: %v\n", r.URL, r.StatusCode, r.Latency)
	}
}
