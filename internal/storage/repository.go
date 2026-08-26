package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *DB
}

func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSession(ctx context.Context, session *Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.StartedAt.IsZero() {
		session.StartedAt = time.Now().UTC()
	}
	if session.Status == "" {
		session.Status = SessionStatusRunning
	}

	query := `INSERT INTO sessions (id, command, status, started_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, session.ID, session.Command, string(session.Status), session.StartedAt)
	return err
}

func (r *Repository) EndSession(ctx context.Context, sessionID string, status SessionStatus) error {
	now := time.Now().UTC()
	query := `UPDATE sessions SET status = ?, ended_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, string(status), now, sessionID)
	return err
}

func (r *Repository) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	query := `SELECT id, command, status, started_at, ended_at FROM sessions WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, sessionID)

	var s Session
	var endedAt sql.NullTime
	var statusStr string

	if err := row.Scan(&s.ID, &s.Command, &statusStr, &s.StartedAt, &endedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.Status = SessionStatus(statusStr)
	if endedAt.Valid {
		s.EndedAt = &endedAt.Time
	}
	return &s, nil
}

func (r *Repository) ListSessions(ctx context.Context, limit int) ([]Session, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, command, status, started_at, ended_at FROM sessions ORDER BY started_at DESC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		var endedAt sql.NullTime
		var statusStr string
		if err := rows.Scan(&s.ID, &s.Command, &statusStr, &s.StartedAt, &endedAt); err != nil {
			return nil, err
		}
		s.Status = SessionStatus(statusStr)
		if endedAt.Valid {
			s.EndedAt = &endedAt.Time
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *Repository) SaveMessage(ctx context.Context, msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now().UTC()
	}

	query := `INSERT INTO messages (id, session_id, direction, method, rpc_id, is_error, payload, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID, msg.SessionID, string(msg.Direction), msg.Method, msg.RPCID, msg.IsError, msg.Payload, msg.CreatedAt)
	return err
}

func (r *Repository) SaveToolCall(ctx context.Context, call *ToolCall) error {
	if call.ID == "" {
		call.ID = uuid.New().String()
	}
	if call.CreatedAt.IsZero() {
		call.CreatedAt = time.Now().UTC()
	}
	if call.Status == "" {
		call.Status = ToolCallStatusPending
	}

	query := `INSERT INTO tool_calls (id, session_id, rpc_id, tool_name, arguments, result, is_error, error_message, latency_ms, status, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		call.ID, call.SessionID, call.RPCID, call.ToolName, call.Arguments, call.Result, call.IsError, call.ErrorMessage, call.LatencyMs, string(call.Status), call.CreatedAt)
	return err
}

func (r *Repository) UpdateToolCallResult(ctx context.Context, sessionID, rpcID string, result string, isError bool, errorMsg string, latencyMs int64, completedAt time.Time) (*ToolCall, error) {
	status := ToolCallStatusCompleted
	if isError {
		status = ToolCallStatusFailed
	}

	var errMsgPtr *string
	if errorMsg != "" {
		errMsgPtr = &errorMsg
	}

	query := `UPDATE tool_calls 
	          SET result = ?, is_error = ?, error_message = ?, latency_ms = ?, status = ?, completed_at = ?
	          WHERE session_id = ? AND rpc_id = ? AND status = 'pending'`
	res, err := r.db.ExecContext(ctx, query, result, isError, errMsgPtr, latencyMs, string(status), completedAt, sessionID, rpcID)
	if err != nil {
		return nil, fmt.Errorf("failed to update tool call: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, nil
	}

	// Fetch updated call
	fetchQuery := `SELECT id, session_id, rpc_id, tool_name, arguments, result, is_error, error_message, latency_ms, status, created_at, completed_at
	               FROM tool_calls WHERE session_id = ? AND rpc_id = ? ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, fetchQuery, sessionID, rpcID)

	var tc ToolCall
	var resStr, errStr sql.NullString
	var statusStr string
	var compAt sql.NullTime

	if err := row.Scan(&tc.ID, &tc.SessionID, &tc.RPCID, &tc.ToolName, &tc.Arguments, &resStr, &tc.IsError, &errStr, &tc.LatencyMs, &statusStr, &tc.CreatedAt, &compAt); err != nil {
		return nil, err
	}
	tc.Status = ToolCallStatus(statusStr)
	if resStr.Valid {
		tc.Result = &resStr.String
	}
	if errStr.Valid {
		tc.ErrorMessage = &errStr.String
	}
	if compAt.Valid {
		tc.CompletedAt = &compAt.Time
	}

	return &tc, nil
}

func (r *Repository) SaveFailure(ctx context.Context, failure *Failure) error {
	if failure.ID == "" {
		failure.ID = uuid.New().String()
	}
	if failure.CreatedAt.IsZero() {
		failure.CreatedAt = time.Now().UTC()
	}
	if failure.Severity == "" {
		failure.Severity = SeverityWarning
	}

	query := `INSERT INTO failures (id, session_id, tool_call_id, failure_type, description, severity, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		failure.ID, failure.SessionID, failure.ToolCallID, string(failure.FailureType), failure.Description, string(failure.Severity), failure.CreatedAt)
	return err
}

func (r *Repository) GetRecentTraces(ctx context.Context, sessionID string, limit int) ([]ToolCall, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, session_id, rpc_id, tool_name, arguments, result, is_error, error_message, latency_ms, status, created_at, completed_at
	          FROM tool_calls 
	          WHERE session_id = ? 
	          ORDER BY created_at DESC 
	          LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var traces []ToolCall
	for rows.Next() {
		var tc ToolCall
		var resStr, errStr sql.NullString
		var statusStr string
		var compAt sql.NullTime

		if err := rows.Scan(&tc.ID, &tc.SessionID, &tc.RPCID, &tc.ToolName, &tc.Arguments, &resStr, &tc.IsError, &errStr, &tc.LatencyMs, &statusStr, &tc.CreatedAt, &compAt); err != nil {
			return nil, err
		}
		tc.Status = ToolCallStatus(statusStr)
		if resStr.Valid {
			tc.Result = &resStr.String
		}
		if errStr.Valid {
			tc.ErrorMessage = &errStr.String
		}
		if compAt.Valid {
			tc.CompletedAt = &compAt.Time
		}
		traces = append(traces, tc)
	}
	return traces, rows.Err()
}

func (r *Repository) GetFailures(ctx context.Context, sessionID string, limit int) ([]Failure, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, session_id, tool_call_id, failure_type, description, severity, created_at
	          FROM failures 
	          WHERE session_id = ? 
	          ORDER BY created_at DESC 
	          LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var failures []Failure
	for rows.Next() {
		var f Failure
		var toolCallID sql.NullString
		var failType, sev string

		if err := rows.Scan(&f.ID, &f.SessionID, &toolCallID, &failType, &f.Description, &sev, &f.CreatedAt); err != nil {
			return nil, err
		}
		if toolCallID.Valid {
			f.ToolCallID = &toolCallID.String
		}
		f.FailureType = FailureType(failType)
		f.Severity = Severity(sev)
		failures = append(failures, f)
	}
	return failures, rows.Err()
}

func (r *Repository) GetSessionStats(ctx context.Context, sessionID string) (*SessionStats, error) {
	query := `
	SELECT 
		COUNT(id) as total_calls,
		SUM(CASE WHEN is_error = 1 THEN 1 ELSE 0 END) as error_calls,
		COALESCE(AVG(latency_ms), 0) as avg_latency
	FROM tool_calls 
	WHERE session_id = ?`

	var totalCalls, errorCalls sql.NullInt64
	var avgLatency sql.NullFloat64

	row := r.db.QueryRowContext(ctx, query, sessionID)
	if err := row.Scan(&totalCalls, &errorCalls, &avgLatency); err != nil {
		return nil, err
	}

	failCountQuery := `SELECT COUNT(id) FROM failures WHERE session_id = ?`
	var failCount int64
	_ = r.db.QueryRowContext(ctx, failCountQuery, sessionID).Scan(&failCount)

	tCalls := totalCalls.Int64
	eCalls := errorCalls.Int64
	successRate := 100.0
	if tCalls > 0 {
		successRate = float64(tCalls-eCalls) / float64(tCalls) * 100.0
	}

	return &SessionStats{
		SessionID:    sessionID,
		TotalCalls:   tCalls,
		ErrorCalls:   eCalls,
		SuccessRate:  successRate,
		AvgLatencyMs: int64(avgLatency.Float64),
		FailureCount: failCount,
	}, nil
}
