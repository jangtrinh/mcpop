package proxy

import (
	"encoding/json"
	"time"
)

// RawRPC represents a generic JSON-RPC 2.0 message
type RawRPC struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error object
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ToolCallParams represents params for "tools/call"
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ToolCallResult represents standard result structure in MCP tools/call
type ToolCallResult struct {
	Content []ContentItem   `json:"content,omitempty"`
	IsError bool            `json:"isError,omitempty"`
}

// ContentItem represents content item in MCP
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// InFlightCall tracks a pending tools/call request
type InFlightCall struct {
	RPCID     string
	ToolName  string
	Arguments string
	SentAt    time.Time
}

// ProxyEvent is passed to async workers
type ProxyEvent struct {
	Type        EventType
	SessionID   string
	Direction   string
	Method      string
	RPCID       string
	Payload     string
	ToolName    string
	Arguments   string
	Result      string
	IsError     bool
	ErrorMsg    string
	LatencyMs   int64
	Timestamp   time.Time
}

type EventType string

const (
	EventMessage  EventType = "message"
	EventToolReq  EventType = "tool_req"
	EventToolResp EventType = "tool_resp"
)

// Helper to stringify JSON-RPC ID (number or string)
func FormatRPCID(rawID json.RawMessage) string {
	if len(rawID) == 0 {
		return ""
	}
	var strID string
	if err := json.Unmarshal(rawID, &strID); err == nil {
		return strID
	}
	return string(rawID)
}
