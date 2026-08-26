package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jangtrinh/mcpop/internal/storage"
)

type Server struct {
	repo       *storage.Repository
	hub        *SSEHub
	port       int
	bind       string
	httpServer *http.Server
}

func NewServer(repo *storage.Repository, hub *SSEHub, port int) *Server {
	if port <= 0 {
		port = 4040
	}
	if hub == nil {
		hub = NewSSEHub()
	}
	return &Server{
		repo: repo,
		hub:  hub,
		port: port,
		bind: "127.0.0.1",
	}
}

func (s *Server) GetHub() *SSEHub {
	return s.hub
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) Bind() string {
	if strings.TrimSpace(s.bind) == "" {
		return "127.0.0.1"
	}
	return s.bind
}

func (s *Server) SetBind(bind string) {
	s.bind = strings.TrimSpace(bind)
}

func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Bind(), s.port)
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/sessions/", s.handleSessionSubRoutes)
	mux.HandleFunc("/api/events", s.handleSSE)
	mux.HandleFunc("/api/replay", s.handleReplay)
	return s.localOnlyMiddleware(mux)
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         s.Addr(),
		Handler:      s.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // Keep 0 for SSE long-lived connections
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) localOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")

		if origin := r.Header.Get("Origin"); origin != "" && !localhostOrigin(origin) {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func localhostOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := u.Hostname()
	return host == "127.0.0.1" || host == "localhost" || host == "::1"
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	DashboardHandler()(w, r)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	sessions, err := s.repo.ListSessions(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sessions == nil {
		sessions = []storage.Session{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sessions)
}

func (s *Server) handleSessionSubRoutes(w http.ResponseWriter, r *http.Request) {
	// Pattern: /api/sessions/{id}/{sub}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	sessionID := parts[0]
	subRoute := ""
	if len(parts) > 1 {
		subRoute = parts[1]
	}

	switch subRoute {
	case "":
		// GET /api/sessions/{id}
		session, err := s.repo.GetSession(r.Context(), sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if session == nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(session)

	case "traces":
		// GET /api/sessions/{id}/traces
		limit := 100
		if l := r.URL.Query().Get("limit"); l != "" {
			if val, err := strconv.Atoi(l); err == nil && val > 0 {
				limit = val
			}
		}

		traces, err := s.repo.GetRecentTraces(r.Context(), sessionID, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if traces == nil {
			traces = []storage.ToolCall{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(traces)

	case "failures":
		// GET /api/sessions/{id}/failures
		failures, err := s.repo.GetFailures(r.Context(), sessionID, 100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if failures == nil {
			failures = []storage.Failure{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(failures)

	case "stats":
		// GET /api/sessions/{id}/stats
		stats, err := s.repo.GetSessionStats(r.Context(), sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)

	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	clientChan := make(chan []byte, 128)
	s.hub.Register(clientChan)
	defer s.hub.Unregister(clientChan)

	// Send initial ping
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			_, _ = w.Write(msg)
			flusher.Flush()
		}
	}
}

func (s *Server) handleReplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.SessionID) == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	session, err := s.repo.GetSession(r.Context(), req.SessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.NotFound(w, r)
		return
	}

	// Never trust a command from the browser. Replay the recorded session command only.
	req.Command = session.Command

	resp, err := ExecuteReplay(r.Context(), req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
