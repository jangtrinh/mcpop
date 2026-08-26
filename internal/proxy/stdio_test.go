package proxy_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mcpop/internal/proxy"
	"mcpop/internal/storage"
)

func TestStdioProxyWithMockServer(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	ctx := context.Background()

	sessionID := "test-proxy-session"
	if err := repo.CreateSession(ctx, &storage.Session{
		ID:      sessionID,
		Command: "python3 test/mock_server.py",
	}); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	p := proxy.NewStdioProxy(repo, sessionID, "python3", []string{"../../test/mock_server.py"}, false)

	// Prepare simulated JSON-RPC messages from AI Client
	inputMessages := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"calculate","arguments":{"expr":"10*5"}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"fail_tool","arguments":{}}}`,
	}

	inBuf := bytes.NewBufferString(strings.Join(inputMessages, "\n") + "\n")
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	p.Stdin = inBuf
	p.Stdout = &outBuf
	p.Stderr = &errBuf

	// Run proxy
	if err := p.Run(ctx); err != nil {
		t.Fatalf("proxy run failed: %v", err)
	}

	// Verify stdout received valid JSON-RPC responses
	outStr := outBuf.String()
	if !strings.Contains(outStr, "mock-mcp-server") {
		t.Errorf("expected initialize response in stdout, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Result: 50") {
		t.Errorf("expected calculate response in stdout, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Database connection timeout") {
		t.Errorf("expected fail_tool response in stdout, got: %s", outStr)
	}

	// Allow worker to finish storing records
	time.Sleep(50 * time.Millisecond)

	// Verify database traces
	traces, err := repo.GetRecentTraces(ctx, sessionID, 10)
	if err != nil {
		t.Fatalf("failed to get traces: %v", err)
	}
	if len(traces) != 2 {
		t.Fatalf("expected 2 tool call traces, got %d", len(traces))
	}

	// Verify calculate trace
	var calcTrace, failTrace *storage.ToolCall
	for i := range traces {
		if traces[i].ToolName == "calculate" {
			calcTrace = &traces[i]
		} else if traces[i].ToolName == "fail_tool" {
			failTrace = &traces[i]
		}
	}

	if calcTrace == nil {
		t.Fatal("expected calculate trace, not found")
	}
	if calcTrace.IsError {
		t.Errorf("calculate should not be error")
	}
	if calcTrace.Result == nil || !strings.Contains(*calcTrace.Result, "Result: 50") {
		t.Errorf("unexpected calculate result: %v", calcTrace.Result)
	}

	if failTrace == nil {
		t.Fatal("expected fail_tool trace, not found")
	}
	if !failTrace.IsError {
		t.Errorf("fail_tool should be marked as is_error = true")
	}
}
