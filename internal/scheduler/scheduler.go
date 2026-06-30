package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/Harshupadhyay-10/uptimer/internal/checker"
	"github.com/Harshupadhyay-10/uptimer/internal/storage"
)

// Scheduler periodically checks all targets and saves results.
type Scheduler struct {
	store    *storage.Store
	interval time.Duration
	stop     chan struct{}
	wg       sync.WaitGroup
}

// New creates a Scheduler that checks all targets at the given interval.
func New(store *storage.Store, interval time.Duration) *Scheduler {
	return &Scheduler{
		store:    store,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start launches the scheduler in a background goroutine.
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log.Printf("Scheduler started — checking every %v", s.interval)

		// Run once immediately, then on every tick
		s.runChecks()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.runChecks()
			case <-s.stop:
				log.Println("Scheduler stopped")
				return
			}
		}
	}()
}

// Stop signals the scheduler to stop and waits for it to finish.
func (s *Scheduler) Stop() {
	close(s.stop)
	s.wg.Wait()
}

// runChecks fetches all targets and checks them concurrently.
func (s *Scheduler) runChecks() {
	targets, err := s.store.ListTargets()
	if err != nil {
		log.Println("scheduler: failed to list targets:", err)
		return
	}
	if len(targets) == 0 {
		return
	}

	urls := make([]string, len(targets))
	idByURL := make(map[string]int64, len(targets))
	for i, t := range targets {
		urls[i] = t.URL
		idByURL[t.URL] = t.ID
	}

	results := checker.CheckAll(urls, 5*time.Second)

	for _, r := range results {
		id := idByURL[r.URL]
		if err := s.store.SaveCheck(id, r.StatusCode, r.Latency, r.Err, r.CheckedAt); err != nil {
			log.Printf("scheduler: failed to save check for %s: %v", r.URL, err)
			continue
		}
		if r.Err != nil {
			log.Printf("DOWN %s | error: %v", r.URL, r.Err)
		} else {
			log.Printf("UP   %s | status: %d | latency: %v", r.URL, r.StatusCode, r.Latency)
		}
	}
}
