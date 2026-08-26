package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jangtrinh/mcpop/internal/server"
	"github.com/jangtrinh/mcpop/internal/storage"
)

func mockServerPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve source path")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "test", "mock_server.py")
}

func setupTestServer(t *testing.T) (*server.Server, *storage.Repository, string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := storage.NewRepository(db)
	hub := server.NewSSEHub()
	srv := server.NewServer(repo, hub, 0)

	sessionID := "test-session-api"
	if err := repo.CreateSession(context.Background(), &storage.Session{
		ID:      sessionID,
		Command: "python3 " + mockServerPath(t),
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
	body := w.Body.String()
	if !strings.Contains(body, "MCPOp") {
		t.Errorf("expected dashboard HTML with MCPOp title, got: %s", body)
	}
	if !strings.Contains(body, "session_id: currentSessionId") {
		t.Errorf("expected replay to send session_id instead of a client-supplied command")
	}
}

func TestServerDefaultsToLocalhost(t *testing.T) {
	srv := server.NewServer(nil, nil, 0)
	if srv.Bind() != "127.0.0.1" {
		t.Fatalf("expected default bind 127.0.0.1, got %s", srv.Bind())
	}
	if srv.Addr() != "127.0.0.1:4040" {
		t.Fatalf("expected default addr 127.0.0.1:4040, got %s", srv.Addr())
	}
}

func TestServerRESTEndpoints(t *testing.T) {
	srv, repo, sessionID := setupTestServer(t)
	ctx := context.Background()
	handler := srv.Handler()

	_ = repo.SaveToolCall(ctx, &storage.ToolCall{
		SessionID: sessionID,
		RPCID:     "100",
		ToolName:  "calculate",
		Arguments: `{"expr":"1+1"}`,
		Status:    storage.ToolCallStatusCompleted,
		LatencyMs: 15,
	})

	req := httptest.NewRequest("GET", "/api/sessions", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/sessions status %d: %s", w.Code, w.Body.String())
	}

	var sessions []storage.Session
	if err := json.Unmarshal(w.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("failed to decode sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	traceReq := httptest.NewRequest("GET", "/api/sessions/"+sessionID+"/traces", nil)
	traceW := httptest.NewRecorder()
	handler.ServeHTTP(traceW, traceReq)
	if traceW.Code != http.StatusOK {
		t.Fatalf("GET traces status %d: %s", traceW.Code, traceW.Body.String())
	}
	var traces []storage.ToolCall
	if err := json.Unmarshal(traceW.Body.Bytes(), &traces); err != nil {
		t.Fatalf("failed to decode traces: %v", err)
	}
	if len(traces) != 1 || traces[0].ToolName != "calculate" {
		t.Fatalf("unexpected traces: %+v", traces)
	}
}

func TestReplayUsesSessionCommandNotRequestBody(t *testing.T) {
	srv, _, sessionID := setupTestServer(t)
	handler := srv.Handler()

	body := map[string]interface{}{
		"session_id": sessionID,
		"command":    "touch /tmp/mcpop-should-not-run",
		"tool_name":  "calculate",
		"arguments":  map[string]interface{}{"expr": "30 * 2"},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/replay", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("replay status %d: %s", w.Code, w.Body.String())
	}

	var resp server.ReplayResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode replay: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected replay success, got error: %s", resp.Error)
	}
}

func TestReplayRejectsMissingSession(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	handler := srv.Handler()

	body := `{"command":"python3","tool_name":"calculate","arguments":{"expr":"1"}}`
	req := httptest.NewRequest("POST", "/api/replay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOriginPolicy(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	handler := srv.Handler()

	evil := httptest.NewRequest("GET", "/api/sessions", nil)
	evil.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, evil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden origin, got %d", w.Code)
	}

	local := httptest.NewRequest("GET", "/api/sessions", nil)
	local.Header.Set("Origin", "http://127.0.0.1:4040")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, local)
	if w.Code != http.StatusOK {
		t.Fatalf("expected localhost origin allowed, got %d: %s", w.Code, w.Body.String())
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
