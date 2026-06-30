package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Harshupadhyay-10/uptimer/internal/checker"
	"github.com/Harshupadhyay-10/uptimer/internal/storage"
)

// Server holds the dependencies for our API.
type Server struct {
	store *storage.Store
}

// New creates a new Server and registers all routes.
func New(store *storage.Store) *Server {
	return &Server{store: store}
}

// RegisterRoutes wires up all HTTP routes on the given mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /targets", s.addTarget)
	mux.HandleFunc("GET /targets", s.listTargets)
	mux.HandleFunc("GET /targets/{id}/check", s.checkTarget)
	mux.HandleFunc("GET /targets/{id}/history", s.targetHistory)
}

// writeJSON is a helper that sends any value as a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError sends a JSON error message.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// addTarget handles POST /targets
// Body: {"url": "https://example.com"}
func (s *Server) addTarget(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}
	if !strings.HasPrefix(body.URL, "http") {
		writeError(w, http.StatusBadRequest, "url must start with http or https")
		return
	}

	id, err := s.store.AddTarget(body.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add target: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":  id,
		"url": body.URL,
	})
}

// listTargets handles GET /targets
func (s *Server) listTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := s.store.ListTargets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list targets")
		return
	}
	writeJSON(w, http.StatusOK, targets)
}

// checkTarget handles GET /targets/{id}/check
// Runs a live check against the target right now.
func (s *Server) checkTarget(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid target id")
		return
	}

	targets, err := s.store.ListTargets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch targets")
		return
	}

	var targetURL string
	for _, t := range targets {
		if t.ID == id {
			targetURL = t.URL
			break
		}
	}
	if targetURL == "" {
		writeError(w, http.StatusNotFound, "target not found")
		return
	}

	result := checker.Check(targetURL, 5*time.Second)

	errMsg := ""
	if result.Err != nil {
		errMsg = result.Err.Error()
	}

	if saveErr := s.store.SaveCheck(id, result.StatusCode, result.Latency, result.Err, result.CheckedAt); saveErr != nil {
		writeError(w, http.StatusInternalServerError, "failed to save check")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"url":        targetURL,
		"status":     result.StatusCode,
		"latency_ms": result.Latency.Milliseconds(),
		"error":      errMsg,
		"checked_at": result.CheckedAt,
	})
}

// targetHistory handles GET /targets/{id}/history
func (s *Server) targetHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid target id")
		return
	}

	checks, err := s.store.ChecksForTarget(id, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch history")
		return
	}

	writeJSON(w, http.StatusOK, checks)
}
