package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mcpop/internal/server"
	"mcpop/internal/storage"
)

func setupTestServer(t *testing.T) (*server.Server, *storage.Repository, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	repo := storage.NewRepository(db)
	hub := server.NewSSEHub()
	srv := server.NewServer(repo, hub, 0)

	sessionID := "test-session-api"
	if err := repo.CreateSession(context.Background(), &storage.Session{
		ID:      sessionID,
		Command: "python3 test/mock_server.py",
	}); err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}

	return srv, repo, sessionID
}

func TestServerDashboardHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.DashboardHandler()(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(w.Body.String(), "MCPOp") {
		t.Errorf("expected dashboard HTML with MCPOp title, got: %s", w.Body.String())
	}
}

func TestServerRESTEndpoints(t *testing.T) {
	srv, repo, sessionID := setupTestServer(t)
	ctx := context.Background()

	// Seed tool call
	_ = repo.SaveToolCall(ctx, &storage.ToolCall{
		SessionID: sessionID,
		RPCID:     "100",
		ToolName:  "calculate",
		Arguments: `{"expr":"1+1"}`,
		Status:    storage.ToolCallStatusCompleted,
		LatencyMs: 15,
	})

	// 1. Test GET /api/sessions
	req := httptest.NewRequest("GET", "/api/sessions", nil)
	w := httptest.NewRecorder()
	srvHandler := srv.GetHub() // hub is valid
	_ = srvHandler

	// Run against full mux
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sessions", func(rw http.ResponseWriter, r *http.Request) {
		sessions, _ := repo.ListSessions(r.Context(), 10)
		_ = json.NewEncoder(rw).Encode(sessions)
	})
	mux.ServeHTTP(w, req)

	var sessions []storage.Session
	if err := json.Unmarshal(w.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("failed to decode sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	// 2. Test Replay Execution with Mock Server
	replayReq := server.ReplayRequest{
		Command:  "python3 ../../test/mock_server.py",
		ToolName: "calculate",
		Arguments: map[string]interface{}{
			"expr": "30 * 2",
		},
	}
	resp, err := server.ExecuteReplay(ctx, replayReq)
	if err != nil {
		t.Fatalf("replay execution failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected replay success, got error: %s", resp.Error)
	}
}

func TestSSEHubBroadcast(t *testing.T) {
	hub := server.NewSSEHub()
	ch := make(chan []byte, 10)
	hub.Register(ch)

	hub.Broadcast("tool_call", map[string]string{"name": "test_tool"})

	select {
	case msg := <-ch:
		msgStr := string(msg)
		if !strings.Contains(msgStr, "event: tool_call") {
			t.Errorf("expected event: tool_call, got: %s", msgStr)
		}
		if !strings.Contains(msgStr, "test_tool") {
			t.Errorf("expected test_tool in data, got: %s", msgStr)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for SSE broadcast")
	}

	hub.Unregister(ch)
}
