package server

import (
	"context"
	"strings"
	"testing"
)

func TestSplitCommandRejectsShellMetacharacters(t *testing.T) {
	_, _, err := splitCommand("python3 server.py; rm -rf /")
	if err == nil {
		t.Fatal("expected shell metacharacter rejection")
	}
	if !strings.Contains(err.Error(), "shell metacharacters") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSplitCommandRejectsEmpty(t *testing.T) {
	_, _, err := splitCommand("   ")
	if err == nil {
		t.Fatal("expected empty command rejection")
	}
}

func TestSplitCommandParsesSimpleArgs(t *testing.T) {
	name, args, err := splitCommand("python3 test/mock_server.py")
	if err != nil {
		t.Fatal(err)
	}
	if name != "python3" {
		t.Fatalf("name=%s", name)
	}
	if len(args) != 1 || args[0] != "test/mock_server.py" {
		t.Fatalf("args=%v", args)
	}
}

func TestSplitCommandParsesQuotedArgs(t *testing.T) {
	name, args, err := splitCommand(`python3 "/path with spaces/server.py" 'arg with spaces'`)
	if err != nil {
		t.Fatal(err)
	}
	if name != "python3" {
		t.Fatalf("name=%s", name)
	}
	if len(args) != 2 || args[0] != "/path with spaces/server.py" || args[1] != "arg with spaces" {
		t.Fatalf("args=%v", args)
	}
}

func TestExecuteReplayRequiresToolName(t *testing.T) {
	_, err := ExecuteReplay(context.Background(), ReplayRequest{Command: "python3 server.py"})
	if err == nil || !strings.Contains(err.Error(), "tool_name") {
		t.Fatalf("expected tool_name error, got %v", err)
	}
}
