package analyzer_test

import (
	"context"
	"path/filepath"
	"testing"

	"mcpop/internal/analyzer"
	"mcpop/internal/storage"
)

func setupTestRepo(t *testing.T) (*storage.Repository, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	repo := storage.NewRepository(db)
	sessionID := "test-analyzer-session"
	if err := repo.CreateSession(context.Background(), &storage.Session{
		ID:      sessionID,
		Command: "mock-command",
	}); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	return repo, sessionID
}

func TestHeuristics_SchemaMismatch(t *testing.T) {
	repo, sessionID := setupTestRepo(t)
	engine := analyzer.NewEngine(repo, 5000)
	ctx := context.Background()

	// 1. Feed tools/list schema
	toolsListJSON := `{"tools": [
		{
			"name": "search_db",
			"description": "Search database",
			"inputSchema": {
				"type": "object",
				"properties": {
					"query": {"type": "string"},
					"limit": {"type": "number"}
				},
				"required": ["query"]
			}
		}
	]}`
	engine.ProcessToolsListResponse(sessionID, []byte(toolsListJSON))

	// 2. Test missing required argument
	failures := engine.CheckToolCallRequest(ctx, sessionID, "call-1", "search_db", `{"limit": 10}`)
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure for missing 'query', got %d", len(failures))
	}
	if failures[0].FailureType != storage.FailureTypeSchemaMismatch {
		t.Errorf("expected FailureTypeSchemaMismatch, got %s", failures[0].FailureType)
	}

	// 3. Test wrong type (query as number instead of string)
	failures = engine.CheckToolCallRequest(ctx, sessionID, "call-2", "search_db", `{"query": 12345, "limit": "not-a-number"}`)
	if len(failures) != 2 {
		t.Fatalf("expected 2 type mismatch failures, got %d", len(failures))
	}

	// 4. Test valid arguments
	failures = engine.CheckToolCallRequest(ctx, sessionID, "call-3", "search_db", `{"query": "golang", "limit": 10}`)
	if len(failures) != 0 {
		t.Errorf("expected 0 failures for valid args, got %d", len(failures))
	}
}

func TestHeuristics_SlowTool(t *testing.T) {
	repo, sessionID := setupTestRepo(t)
	engine := analyzer.NewEngine(repo, 2000) // threshold = 2000ms
	ctx := context.Background()

	failures := engine.CheckToolCallResponse(ctx, sessionID, "call-slow", "generate_embedding", `{"text":"hello"}`, false, "", 2500)
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure for slow tool, got %d", len(failures))
	}
	if failures[0].FailureType != storage.FailureTypeSlowTool {
		t.Errorf("expected FailureTypeSlowTool, got %s", failures[0].FailureType)
	}
}

func TestHeuristics_LoopDetection(t *testing.T) {
	repo, sessionID := setupTestRepo(t)
	engine := analyzer.NewEngine(repo, 5000)
	ctx := context.Background()

	// First failure: should not trigger loop yet
	failures1 := engine.CheckToolCallResponse(ctx, sessionID, "call-1", "query_api", `{"endpoint":"/users"}`, true, "500 Internal Server Error", 100)
	if len(failures1) != 0 {
		t.Errorf("expected 0 loop failures on first call, got %d", len(failures1))
	}

	// Second consecutive identical failure: should trigger loop!
	failures2 := engine.CheckToolCallResponse(ctx, sessionID, "call-2", "query_api", `{"endpoint":"/users"}`, true, "500 Internal Server Error", 120)
	if len(failures2) != 1 {
		t.Fatalf("expected 1 loop failure on 2nd consecutive error, got %d", len(failures2))
	}
	if failures2[0].FailureType != storage.FailureTypeLoop {
		t.Errorf("expected FailureTypeLoop, got %s", failures2[0].FailureType)
	}
	if failures2[0].Severity != storage.SeverityCritical {
		t.Errorf("expected SeverityCritical, got %s", failures2[0].Severity)
	}
}
