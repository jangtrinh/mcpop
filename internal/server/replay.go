package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

type ReplayRequest struct {
	SessionID string                 `json:"session_id,omitempty"`
	Command   string                 `json:"command,omitempty"`
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
	if strings.TrimSpace(req.ToolName) == "" {
		return nil, fmt.Errorf("tool_name is required")
	}

	cmdName, cmdArgs, err := splitCommand(req.Command)
	if err != nil {
		return nil, err
	}

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
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)

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
	if err := writeJSONLine(stdin, initReq); err != nil {
		return nil, fmt.Errorf("failed to write initialize: %w", err)
	}
	if _, err := readJSONRPC(scanner, "init-replay"); err != nil {
		return nil, fmt.Errorf("failed to read initialize response: %w", err)
	}

	initialized := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	if err := writeJSONLine(stdin, initialized); err != nil {
		return nil, fmt.Errorf("failed to write initialized notification: %w", err)
	}

	callReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "call-replay",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      req.ToolName,
			"arguments": req.Arguments,
		},
	}
	start := time.Now()
	if err := writeJSONLine(stdin, callReq); err != nil {
		return nil, fmt.Errorf("failed to write tools/call: %w", err)
	}

	respLine, err := readJSONRPC(scanner, "call-replay")
	if err != nil {
		return nil, fmt.Errorf("failed to read tools/call response: %w", err)
	}
	latencyMs := time.Since(start).Milliseconds()

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

func splitCommand(command string) (string, []string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", nil, fmt.Errorf("command is required")
	}

	var parts []string
	var current strings.Builder
	inQuote := rune(0)
	escaped := false

	for _, r := range command {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}
		if r == '"' || r == '\'' {
			inQuote = r
			continue
		}
		if unicode.IsSpace(r) {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}
		if strings.ContainsRune(";|&$`<>(){}", r) {
			return "", nil, fmt.Errorf("command contains shell metacharacters")
		}
		current.WriteRune(r)
	}

	if inQuote != 0 {
		return "", nil, fmt.Errorf("unclosed quote in command")
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	if len(parts) == 0 {
		return "", nil, fmt.Errorf("command is required")
	}

	for _, part := range parts {
		if strings.Contains(part, "..") {
			return "", nil, fmt.Errorf("command contains parent-directory path")
		}
	}

	return parts[0], parts[1:], nil
}

func writeJSONLine(w interface{ Write([]byte) (int, error) }, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

func readJSONRPC(scanner *bufio.Scanner, wantID string) ([]byte, error) {
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 || line[0] != '{' {
			continue
		}

		var msg struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}
		if len(msg.ID) == 0 {
			continue
		}
		if !rpcIDEqual(msg.ID, wantID) {
			continue
		}

		copied := make([]byte, len(line))
		copy(copied, line)
		return copied, nil
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("server closed stdout before %s response", wantID)
}

func rpcIDEqual(raw json.RawMessage, want string) bool {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return false
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString == want
	}
	var asNumber json.Number
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return string(asNumber) == want
	}
	trimmed := strings.TrimFunc(string(raw), func(r rune) bool {
		return unicode.IsSpace(r) || r == '"'
	})
	return trimmed == want
}
