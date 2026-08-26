package proxy

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"mcpop/internal/analyzer"
	"mcpop/internal/server"
	"mcpop/internal/storage"
)

type StdioProxy struct {
	repo       *storage.Repository
	analyzer   *analyzer.Engine
	hub        *server.SSEHub
	sessionID  string
	cmdName    string
	cmdArgs    []string
	inFlight   sync.Map // map[string]*InFlightCall
	eventChan  chan ProxyEvent
	workerDone chan struct{}
	debug      bool

	// Custom I/O overrides for testing
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func NewStdioProxy(repo *storage.Repository, sessionID string, cmdName string, cmdArgs []string, debug bool) *StdioProxy {
	return &StdioProxy{
		repo:       repo,
		sessionID:  sessionID,
		cmdName:    cmdName,
		cmdArgs:    cmdArgs,
		eventChan:  make(chan ProxyEvent, 2048),
		workerDone: make(chan struct{}),
		debug:      debug,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	}
}

func (p *StdioProxy) SetAnalyzer(a *analyzer.Engine) {
	p.analyzer = a
}

func (p *StdioProxy) SetSSEHub(hub *server.SSEHub) {
	p.hub = hub
}

// Run executes the child process and proxies stdin/stdout/stderr
func (p *StdioProxy) Run(ctx context.Context) error {
	// Start async storage worker
	go p.storageWorker()

	// Prepare child process
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd := exec.CommandContext(childCtx, p.cmdName, p.cmdArgs...)

	childStdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer childStdin.Close()

	childStdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	childStderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Forward OS signals to child process
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		select {
		case sig := <-sigChan:
			if cmd.Process != nil {
				_ = cmd.Process.Signal(sig)
			}
		case <-childCtx.Done():
		}
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command '%s': %w", p.cmdName, err)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	// 1. Client -> Server (Stdin)
	go func() {
		defer wg.Done()
		defer childStdin.Close()
		p.pipeClientToServer(p.Stdin, childStdin)
	}()

	// 2. Server -> Client (Stdout)
	go func() {
		defer wg.Done()
		p.pipeServerToClient(childStdout, p.Stdout)
	}()

	// 3. Child Stderr -> Client Stderr
	go func() {
		defer wg.Done()
		_, _ = io.Copy(p.Stderr, childStderr)
	}()

	// Wait for I/O forwarding to finish
	wg.Wait()

	// Wait for process exit
	cmdErr := cmd.Wait()

	// Flush and close storage worker
	close(p.eventChan)
	<-p.workerDone

	// Update session end status
	sessionStatus := storage.SessionStatusCompleted
	if cmdErr != nil {
		sessionStatus = storage.SessionStatusFailed
	}
	_ = p.repo.EndSession(context.Background(), p.sessionID, sessionStatus)

	if p.hub != nil {
		p.hub.Broadcast("session_ended", map[string]string{
			"session_id": p.sessionID,
			"status":     string(sessionStatus),
		})
	}

	return cmdErr
}

func (p *StdioProxy) pipeClientToServer(src io.Reader, dst io.Writer) {
	scanner := bufio.NewScanner(src)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		// Forward immediately
		_, _ = dst.Write(line)
		_, _ = dst.Write([]byte("\n"))

		if len(line) == 0 {
			continue
		}

		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)
		p.handleClientMessage(lineCopy)
	}
}

func (p *StdioProxy) pipeServerToClient(src io.Reader, dst io.Writer) {
	scanner := bufio.NewScanner(src)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		// Forward immediately
		_, _ = dst.Write(line)
		_, _ = dst.Write([]byte("\n"))

		if len(line) == 0 {
			continue
		}

		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)
		p.handleServerMessage(lineCopy)
	}
}

func (p *StdioProxy) handleClientMessage(data []byte) {
	var rpc RawRPC
	if err := json.Unmarshal(data, &rpc); err != nil {
		if p.debug {
			log.Printf("[mcpop] non-json-rpc client message: %s", string(data))
		}
		return
	}

	rpcID := FormatRPCID(rpc.ID)
	now := time.Now().UTC()

	// Emit raw message event
	p.emitEvent(ProxyEvent{
		Type:      EventMessage,
		SessionID: p.sessionID,
		Direction: string(storage.DirectionClientToServer),
		Method:    rpc.Method,
		RPCID:     rpcID,
		Payload:   string(data),
		Timestamp: now,
	})

	// Intercept tools/call request
	if rpc.Method == "tools/call" {
		var params ToolCallParams
		if err := json.Unmarshal(rpc.Params, &params); err == nil {
			argsStr := "{}"
			if len(params.Arguments) > 0 {
				argsStr = string(params.Arguments)
			}

			if rpcID != "" {
				p.inFlight.Store(rpcID, &InFlightCall{
					RPCID:     rpcID,
					ToolName:  params.Name,
					Arguments: argsStr,
					SentAt:    now,
				})
			}

			p.emitEvent(ProxyEvent{
				Type:      EventToolReq,
				SessionID: p.sessionID,
				RPCID:     rpcID,
				ToolName:  params.Name,
				Arguments: argsStr,
				Timestamp: now,
			})
		}
	}
}

func (p *StdioProxy) handleServerMessage(data []byte) {
	var rpc RawRPC
	if err := json.Unmarshal(data, &rpc); err != nil {
		if p.debug {
			log.Printf("[mcpop] non-json-rpc server message: %s", string(data))
		}
		return
	}

	rpcID := FormatRPCID(rpc.ID)
	now := time.Now().UTC()
	isError := rpc.Error != nil

	p.emitEvent(ProxyEvent{
		Type:      EventMessage,
		SessionID: p.sessionID,
		Direction: string(storage.DirectionServerToClient),
		Method:    rpc.Method,
		RPCID:     rpcID,
		IsError:   isError,
		Payload:   string(data),
		Timestamp: now,
	})

	// Check if this is a response to tools/list (cache schema)
	if len(rpc.Result) > 0 && p.analyzer != nil {
		var listCheck struct {
			Tools []analyzer.ToolSchema `json:"tools"`
		}
		if err := json.Unmarshal(rpc.Result, &listCheck); err == nil && len(listCheck.Tools) > 0 {
			p.analyzer.ProcessToolsListResponse(p.sessionID, rpc.Result)
		}
	}

	if rpcID == "" {
		return
	}

	// Check if this matches an in-flight tools/call
	val, ok := p.inFlight.LoadAndDelete(rpcID)
	if !ok {
		return
	}
	inFlight := val.(*InFlightCall)

	latencyMs := now.Sub(inFlight.SentAt).Milliseconds()
	var resultStr string
	var errorMsg string

	if rpc.Error != nil {
		isError = true
		errorMsg = rpc.Error.Message
		resultStr = string(data)
	} else if len(rpc.Result) > 0 {
		resultStr = string(rpc.Result)
		var callResult ToolCallResult
		if err := json.Unmarshal(rpc.Result, &callResult); err == nil {
			if callResult.IsError {
				isError = true
				if len(callResult.Content) > 0 {
					errorMsg = callResult.Content[0].Text
				}
			}
		}
	}

	p.emitEvent(ProxyEvent{
		Type:      EventToolResp,
		SessionID: p.sessionID,
		RPCID:     rpcID,
		ToolName:  inFlight.ToolName,
		Arguments: inFlight.Arguments,
		Result:    resultStr,
		IsError:   isError,
		ErrorMsg:  errorMsg,
		LatencyMs: latencyMs,
		Timestamp: now,
	})
}

func (p *StdioProxy) emitEvent(evt ProxyEvent) {
	select {
	case p.eventChan <- evt:
	default:
		if p.debug {
			log.Printf("[mcpop] event buffer full, dropping event %s", evt.Type)
		}
	}
}

func (p *StdioProxy) storageWorker() {
	defer close(p.workerDone)
	ctx := context.Background()

	for evt := range p.eventChan {
		switch evt.Type {
		case EventMessage:
			_ = p.repo.SaveMessage(ctx, &storage.Message{
				SessionID: evt.SessionID,
				Direction: storage.Direction(evt.Direction),
				Method:    evt.Method,
				RPCID:     evt.RPCID,
				IsError:   evt.IsError,
				Payload:   evt.Payload,
				CreatedAt: evt.Timestamp,
			})

		case EventToolReq:
			tc := &storage.ToolCall{
				SessionID: evt.SessionID,
				RPCID:     evt.RPCID,
				ToolName:  evt.ToolName,
				Arguments: evt.Arguments,
				Status:    storage.ToolCallStatusPending,
				CreatedAt: evt.Timestamp,
			}
			_ = p.repo.SaveToolCall(ctx, tc)

			// Heuristic Check: Schema Mismatch
			if p.analyzer != nil {
				_ = p.analyzer.CheckToolCallRequest(ctx, evt.SessionID, tc.ID, evt.ToolName, evt.Arguments)
			}

			// Broadcast SSE
			if p.hub != nil {
				p.hub.Broadcast("tool_call", tc)
			}

		case EventToolResp:
			updated, err := p.repo.UpdateToolCallResult(ctx,
				evt.SessionID,
				evt.RPCID,
				evt.Result,
				evt.IsError,
				evt.ErrorMsg,
				evt.LatencyMs,
				evt.Timestamp,
			)
			if err != nil && !errors.Is(err, storage.ErrNotFound) && p.debug {
				log.Printf("[mcpop] failed to update tool call result: %v", err)
			}

			// Heuristic Check: Slow Tool & Loop Detection
			if p.analyzer != nil {
				toolCallID := ""
				if updated != nil {
					toolCallID = updated.ID
				}
				_ = p.analyzer.CheckToolCallResponse(ctx, evt.SessionID, toolCallID, evt.ToolName, evt.Arguments, evt.IsError, evt.ErrorMsg, evt.LatencyMs)
			}

			// Broadcast SSE
			if p.hub != nil && updated != nil {
				p.hub.Broadcast("tool_call", updated)
			}
		}
	}
}
