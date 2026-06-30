# Uptimer

A URL uptime monitor written in Go. It checks whether your URLs are up, stores the results in SQLite, and exposes a REST API to manage everything. A background scheduler re-checks targets automatically so you don't have to.

I built this to get comfortable with Go's concurrency model — goroutines, channels, the fan-out/fan-in pattern — while making something actually useful.

## What it does

- Checks multiple URLs concurrently (goroutines + channels, not sequential loops)
- Stores every check result in SQLite — status code, latency, timestamp, error if any
- REST API to add targets, list them, run a live check, or pull history
- Background scheduler that re-checks all targets every 30 seconds
- Shuts down cleanly on Ctrl+C (no mid-check interruptions)

## Project layout

- main.go    - wires everything together
- internal/
- checker/   - HTTP check logic, concurrent fan-out
- storage/   - SQLite (targets + check history)
- api/       - REST handlers, stdlib router only
- scheduler/ - ackground ticker, graceful stop

No frameworks. Just Go's standard library + `modernc.org/sqlite` (pure Go, no cgo).

## Run it

```bash
git clone https://github.com/Harshupadhyay-10/uptimer
cd uptimer
make run
```

Needs Go 1.22+.

## API

```bash
# Add a URL to monitor
curl -X POST http://localhost:8080/targets \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# List all targets
curl http://localhost:8080/targets

# Run a live check right now
curl http://localhost:8080/targets/1/check

# Pull check history (last 20)
curl http://localhost:8080/targets/1/history
```

## Things I learned building this

Go's concurrency primitives are genuinely different from anything I'd used before. The `select` statement for waiting on multiple channels, using a buffered channel as the collection point for concurrent goroutine results, `sync.WaitGroup` for knowing when background work is done these clicked for me while writing the scheduler and the concurrent checker.

The standard library also goes further than I expected. The 1.22 router handles path parameters (`/targets/{id}/check`) cleanly without needing gorilla/mux or chi.