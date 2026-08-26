package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jangtrinh/mcpop/internal/storage"
)

// ToolSchema represents JSON Schema of a tool in MCP
type ToolSchema struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputSchema *JSONSchemaObject `json:"inputSchema,omitempty"`
}

type JSONSchemaObject struct {
	Type       string                  `json:"type"`
	Properties map[string]JSONProperty `json:"properties,omitempty"`
	Required   []string                `json:"required,omitempty"`
}

type JSONProperty struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type CallHistoryItem struct {
	ToolName  string
	Arguments string
	IsError   bool
	ErrorMsg  string
	CreatedAt time.Time
}

type Engine struct {
	repo            *storage.Repository
	schemaCache     sync.Map // map[string]map[string]ToolSchema (sessionID -> toolName -> ToolSchema)
	historyCache    sync.Map // map[string][]*CallHistoryItem (sessionID -> history)
	slowThresholdMs int64
	onFailure       func(failure *storage.Failure)
	mu              sync.Mutex
}

func NewEngine(repo *storage.Repository, slowThresholdMs int64) *Engine {
	if slowThresholdMs <= 0 {
		slowThresholdMs = 5000
	}
	return &Engine{
		repo:            repo,
		slowThresholdMs: slowThresholdMs,
	}
}

func (e *Engine) SetOnFailureCallback(cb func(failure *storage.Failure)) {
	e.onFailure = cb
}

// ProcessToolsListResponse caches tool schemas from a tools/list response
func (e *Engine) ProcessToolsListResponse(sessionID string, resultPayload []byte) {
	var listResp struct {
		Tools []ToolSchema `json:"tools"`
	}
	if err := json.Unmarshal(resultPayload, &listResp); err != nil {
		return
	}

	schemas := make(map[string]ToolSchema)
	for _, t := range listResp.Tools {
		schemas[t.Name] = t
	}

	e.schemaCache.Store(sessionID, schemas)
}

// CheckToolCallRequest validates schema of a tool call request before/at execution
func (e *Engine) CheckToolCallRequest(ctx context.Context, sessionID string, toolCallID string, toolName string, rawArgs string) []*storage.Failure {
	var failures []*storage.Failure

	val, ok := e.schemaCache.Load(sessionID)
	if !ok {
		return failures
	}
	schemas := val.(map[string]ToolSchema)
	schema, exists := schemas[toolName]
	if !exists || schema.InputSchema == nil {
		return failures
	}

	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(rawArgs), &argsMap); err != nil {
		f := &storage.Failure{
			SessionID:   sessionID,
			ToolCallID:  &toolCallID,
			FailureType: storage.FailureTypeSchemaMismatch,
			Description: fmt.Sprintf("Arguments for tool '%s' are not valid JSON: %s", toolName, rawArgs),
			Severity:    storage.SeverityError,
			CreatedAt:   time.Now().UTC(),
		}
		failures = append(failures, f)
		_ = e.saveFailure(ctx, f)
		return failures
	}

	// 1. Check required arguments
	for _, reqKey := range schema.InputSchema.Required {
		if _, found := argsMap[reqKey]; !found {
			f := &storage.Failure{
				SessionID:   sessionID,
				ToolCallID:  &toolCallID,
				FailureType: storage.FailureTypeSchemaMismatch,
				Description: fmt.Sprintf("Missing required argument '%s' for tool '%s'", reqKey, toolName),
				Severity:    storage.SeverityError,
				CreatedAt:   time.Now().UTC(),
			}
			failures = append(failures, f)
			_ = e.saveFailure(ctx, f)
		}
	}

	// 2. Check property types
	for propName, propDef := range schema.InputSchema.Properties {
		val, found := argsMap[propName]
		if !found || val == nil {
			continue
		}

		typeMismatch := false
		actualType := ""

		switch propDef.Type {
		case "string":
			if _, ok := val.(string); !ok {
				typeMismatch = true
				actualType = fmt.Sprintf("%T", val)
			}
		case "number", "integer":
			if _, ok := val.(float64); !ok {
				typeMismatch = true
				actualType = fmt.Sprintf("%T", val)
			}
		case "boolean":
			if _, ok := val.(bool); !ok {
				typeMismatch = true
				actualType = fmt.Sprintf("%T", val)
			}
		case "array":
			if _, ok := val.([]interface{}); !ok {
				typeMismatch = true
				actualType = fmt.Sprintf("%T", val)
			}
		case "object":
			if _, ok := val.(map[string]interface{}); !ok {
				typeMismatch = true
				actualType = fmt.Sprintf("%T", val)
			}
		}

		if typeMismatch {
			f := &storage.Failure{
				SessionID:   sessionID,
				ToolCallID:  &toolCallID,
				FailureType: storage.FailureTypeSchemaMismatch,
				Description: fmt.Sprintf("Argument '%s' for tool '%s' expected type '%s', got '%s'", propName, toolName, propDef.Type, actualType),
				Severity:    storage.SeverityWarning,
				CreatedAt:   time.Now().UTC(),
			}
			failures = append(failures, f)
			_ = e.saveFailure(ctx, f)
		}
	}

	return failures
}

// CheckToolCallResponse checks for slow tools and loops after tool call completes
func (e *Engine) CheckToolCallResponse(ctx context.Context, sessionID string, toolCallID string, toolName string, rawArgs string, isError bool, errorMsg string, latencyMs int64) []*storage.Failure {
	var failures []*storage.Failure
	now := time.Now().UTC()

	// 1. Check Slow Tool
	if latencyMs >= e.slowThresholdMs {
		f := &storage.Failure{
			SessionID:   sessionID,
			ToolCallID:  &toolCallID,
			FailureType: storage.FailureTypeSlowTool,
			Description: fmt.Sprintf("Tool '%s' execution took %dms (exceeds %dms threshold)", toolName, latencyMs, e.slowThresholdMs),
			Severity:    storage.SeverityWarning,
			CreatedAt:   now,
		}
		failures = append(failures, f)
		_ = e.saveFailure(ctx, f)
	}

	// 2. Track history for Loop detection
	e.mu.Lock()
	var history []*CallHistoryItem
	if val, ok := e.historyCache.Load(sessionID); ok {
		history = val.([]*CallHistoryItem)
	}

	item := &CallHistoryItem{
		ToolName:  toolName,
		Arguments: rawArgs,
		IsError:   isError,
		ErrorMsg:  errorMsg,
		CreatedAt: now,
	}
	history = append(history, item)
	// Keep last 50 items
	if len(history) > 50 {
		history = history[len(history)-50:]
	}
	e.historyCache.Store(sessionID, history)
	e.mu.Unlock()

	// Check Loop pattern: if consecutive calls to the same tool failed with identical arguments
	if isError && len(history) >= 2 {
		consecutiveFailures := 0
		for i := len(history) - 1; i >= 0; i-- {
			h := history[i]
			if h.ToolName == toolName && h.IsError && h.Arguments == rawArgs {
				consecutiveFailures++
			} else {
				break
			}
		}

		if consecutiveFailures >= 2 {
			f := &storage.Failure{
				SessionID:   sessionID,
				ToolCallID:  &toolCallID,
				FailureType: storage.FailureTypeLoop,
				Description: fmt.Sprintf("Agent repeated failed call to '%s' %d times consecutively with identical arguments: %s", toolName, consecutiveFailures, rawArgs),
				Severity:    storage.SeverityCritical,
				CreatedAt:   now,
			}
			failures = append(failures, f)
			_ = e.saveFailure(ctx, f)
		}
	}

	return failures
}

func (e *Engine) saveFailure(ctx context.Context, f *storage.Failure) error {
	if e.repo != nil {
		if err := e.repo.SaveFailure(ctx, f); err != nil {
			return err
		}
	}
	if e.onFailure != nil {
		e.onFailure(f)
	}
	return nil
}
