package storage

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

type SessionStatus string

const (
	SessionStatusRunning   SessionStatus = "running"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusFailed    SessionStatus = "failed"
)

type Direction string

const (
	DirectionClientToServer Direction = "client_to_server"
	DirectionServerToClient Direction = "server_to_client"
)

type ToolCallStatus string

const (
	ToolCallStatusPending   ToolCallStatus = "pending"
	ToolCallStatusCompleted ToolCallStatus = "completed"
	ToolCallStatusFailed    ToolCallStatus = "failed"
)

type FailureType string

const (
	FailureTypeLoop           FailureType = "loop"
	FailureTypeSchemaMismatch FailureType = "schema_mismatch"
	FailureTypeSlowTool       FailureType = "slow_tool"
)

type Severity string

const (
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

type Session struct {
	ID        string        `json:"id"`
	Command   string        `json:"command"`
	Status    SessionStatus `json:"status"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   *time.Time    `json:"ended_at,omitempty"`
}

type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Direction Direction `json:"direction"`
	Method    string    `json:"method,omitempty"`
	RPCID     string    `json:"rpc_id,omitempty"`
	IsError   bool      `json:"is_error"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type ToolCall struct {
	ID           string         `json:"id"`
	SessionID    string         `json:"session_id"`
	RPCID        string         `json:"rpc_id"`
	ToolName     string         `json:"tool_name"`
	Arguments    string         `json:"arguments"`
	Result       *string        `json:"result,omitempty"`
	IsError      bool           `json:"is_error"`
	ErrorMessage *string        `json:"error_message,omitempty"`
	LatencyMs    int64          `json:"latency_ms"`
	Status       ToolCallStatus `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
}

type Failure struct {
	ID          string      `json:"id"`
	SessionID   string      `json:"session_id"`
	ToolCallID  *string     `json:"tool_call_id,omitempty"`
	FailureType FailureType `json:"failure_type"`
	Description string      `json:"description"`
	Severity    Severity    `json:"severity"`
	CreatedAt   time.Time   `json:"created_at"`
}

type SessionStats struct {
	SessionID    string  `json:"session_id"`
	TotalCalls   int64   `json:"total_calls"`
	ErrorCalls   int64   `json:"error_calls"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs int64   `json:"avg_latency_ms"`
	FailureCount int64   `json:"failure_count"`
}
