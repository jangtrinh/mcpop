package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ReplayRequest struct {
	Command   string                 `json:"command"`
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ReplayResponse struct {
	Success   bool        `json:"success"`
	LatencyMs int64       `json:"latency_ms"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
}

func ExecuteReplay(ctx context.Context, req ReplayRequest) (*ReplayResponse, error) {
	if strings.TrimSpace(req.Command) == "" {
		return nil, fmt.Errorf("command is required")
	}
	if strings.TrimSpace(req.ToolName) == "" {
		return nil, fmt.Errorf("tool_name is required")
	}

	parts := strings.Fields(req.Command)
	cmdName := parts[0]
	cmdArgs := parts[1:]

	childCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(childCtx, cmdName, cmdArgs...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin: %w", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	scanner := bufio.NewScanner(stdout)

	// Step 1: Send initialize
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "init-replay",
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "mcpop-replay",
				"version": "1.0.0",
			},
		},
	}
	initBytes, _ := json.Marshal(initReq)
	_, _ = stdin.Write(append(initBytes, '\n'))

	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read initialize response from server")
	}

	// Step 2: Send tools/call
	callReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "call-replay",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      req.ToolName,
			"arguments": req.Arguments,
		},
	}
	callBytes, _ := json.Marshal(callReq)

	start := time.Now()
	_, _ = stdin.Write(append(callBytes, '\n'))

	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read tools/call response from server")
	}
	latencyMs := time.Since(start).Milliseconds()

	respLine := scanner.Bytes()
	var rawResp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respLine, &rawResp); err != nil {
		return &ReplayResponse{
			Success:   false,
			LatencyMs: latencyMs,
			Error:     fmt.Sprintf("Invalid JSON from server: %s", string(respLine)),
		}, nil
	}

	if rawResp.Error != nil {
		return &ReplayResponse{
			Success:   false,
			LatencyMs: latencyMs,
			Error:     rawResp.Error.Message,
		}, nil
	}

	var resultObj interface{}
	_ = json.Unmarshal(rawResp.Result, &resultObj)

	return &ReplayResponse{
		Success:   true,
		LatencyMs: latencyMs,
		Result:    resultObj,
	}, nil
}
