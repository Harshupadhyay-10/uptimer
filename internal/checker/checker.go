package checker

import (
	"net/http"
	"time"
)

// Result holds the outcome of a single health check.
type Result struct {
	URL        string
	StatusCode int
	Latency    time.Duration
	Err        error
	CheckedAt  time.Time
}

// Check performs a single HTTP GET request against the given URL
// and returns a Result describing what happened.
func Check(url string, timeout time.Duration) Result {
	client := http.Client{
		Timeout: timeout,
	}

	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start)

	result := Result{
		URL:       url,
		Latency:   latency,
		CheckedAt: start,
	}

	if err != nil {
		result.Err = err
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	return result
}

// CheckAll runs Check concurrently for every URL in the list
// and returns all results once every check has completed.
func CheckAll(urls []string, timeout time.Duration) []Result {
	resultsChan := make(chan Result, len(urls))

	for _, url := range urls {
		go func(u string) {
			resultsChan <- Check(u, timeout)
		}(url)
	}

	results := make([]Result, 0, len(urls))
	for i := 0; i < len(urls); i++ {
		results = append(results, <-resultsChan)
	}

	return results
}
