package storage_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"mcpop/internal/storage"
)

func TestStorageWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	ctx := context.Background()

	// 1. Test Session creation & retrieval
	session := &storage.Session{
		ID:      "sess-123",
		Command: "python mock_server.py",
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	fetched, err := repo.GetSession(ctx, "sess-123")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if fetched.Command != session.Command {
		t.Errorf("expected command %s, got %s", session.Command, fetched.Command)
	}

	// 2. Test Message save
	msg := &storage.Message{
		SessionID: "sess-123",
		Direction: storage.DirectionClientToServer,
		Method:    "tools/call",
		RPCID:     "1",
		Payload:   `{"jsonrpc":"2.0","id":"1","method":"tools/call"}`,
	}
	if err := repo.SaveMessage(ctx, msg); err != nil {
		t.Fatalf("failed to save message: %v", err)
	}

	// 3. Test ToolCall lifecycle
	call := &storage.ToolCall{
		SessionID: "sess-123",
		RPCID:     "1",
		ToolName:  "calculate",
		Arguments: `{"expr":"2+2"}`,
	}
	if err := repo.SaveToolCall(ctx, call); err != nil {
		t.Fatalf("failed to save tool call: %v", err)
	}

	// Complete tool call
	completed, err := repo.UpdateToolCallResult(ctx, "sess-123", "1", `{"content":[{"type":"text","text":"Result: 4"}]}`, false, "", 42, time.Now().UTC())
	if err != nil {
		t.Fatalf("failed to update tool call: %v", err)
	}
	if completed == nil {
		t.Fatal("expected completed tool call, got nil")
	}
	if completed.LatencyMs != 42 {
		t.Errorf("expected latency 42, got %d", completed.LatencyMs)
	}
	if completed.Status != storage.ToolCallStatusCompleted {
		t.Errorf("expected status completed, got %s", completed.Status)
	}

	// 4. Test GetRecentTraces
	traces, err := repo.GetRecentTraces(ctx, "sess-123", 10)
	if err != nil {
		t.Fatalf("failed to get recent traces: %v", err)
	}
	if len(traces) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(traces))
	}
	if traces[0].ToolName != "calculate" {
		t.Errorf("expected tool name calculate, got %s", traces[0].ToolName)
	}

	// 5. Test Failure save
	failure := &storage.Failure{
		SessionID:   "sess-123",
		ToolCallID:  &traces[0].ID,
		FailureType: storage.FailureTypeSlowTool,
		Description: "Tool execution exceeded 5000ms",
		Severity:    storage.SeverityWarning,
	}
	if err := repo.SaveFailure(ctx, failure); err != nil {
		t.Fatalf("failed to save failure: %v", err)
	}

	// 6. Test End Session
	if err := repo.EndSession(ctx, "sess-123", storage.SessionStatusCompleted); err != nil {
		t.Fatalf("failed to end session: %v", err)
	}
	endedSess, err := repo.GetSession(ctx, "sess-123")
	if err != nil {
		t.Fatalf("failed to get ended session: %v", err)
	}
	if endedSess.Status != storage.SessionStatusCompleted {
		t.Errorf("expected status completed, got %s", endedSess.Status)
	}
	if endedSess.EndedAt == nil {
		t.Error("expected ended_at to be populated")
	}
}
