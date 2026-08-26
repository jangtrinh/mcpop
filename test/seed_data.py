#!/usr/bin/env python3
"""
Seed realistic traffic through MCPOp Stdio Proxy
Spawns: ./bin/mcpop run -- python3 test/mock_server.py
Sends 20+ realistic JSON-RPC requests to generate rich live observability data.
"""
import subprocess
import json
import time
import sys

def main():
    print("🚀 Spawning MCPOp Proxy with Mock Server...")
    proc = subprocess.Popen(
        ["./bin/mcpop", "run", "--", "python3", "test/mock_server.py"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=1
    )

    def send_rpc(msg):
        line = json.dumps(msg) + "\n"
        proc.stdin.write(line)
        proc.stdin.flush()
        resp_line = proc.stdout.readline()
        return json.loads(resp_line) if resp_line else {}

    # 1. Initialize
    print("1. Initializing session...")
    send_rpc({"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05"}})
    time.sleep(0.05)

    # 2. List tools (populates schema cache)
    print("2. Fetching tools/list...")
    send_rpc({"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}})
    time.sleep(0.05)

    # 3. Series of realistic calls
    calls = [
        # Normal traffic
        {"name": "db/query", "arguments": {"sql": "SELECT id, name, status FROM products LIMIT 5"}},
        {"name": "search/web", "arguments": {"query": "model context protocol zero overhead proxy", "max_results": 10}},
        {"name": "fs/read_file", "arguments": {"path": "/etc/mcpop/config.toml"}},
        {"name": "ai/embed_text", "arguments": {"text": "Autonomous Agent Reliability and Observability", "model": "text-embedding-3-small"}},
        {"name": "auth/verify_session", "arguments": {"token": "valid_token_jwt_991823"}},
        {"name": "db/query", "arguments": {"sql": "SELECT count(*) FROM telemetry_events"}},
        
        # Schema Mismatch anomaly (Missing required "model" field)
        {"name": "ai/embed_text", "arguments": {"text": "Vector search payload with missing model property"}},
        
        # More normal traffic
        {"name": "fs/read_file", "arguments": {"path": "/var/log/mcpop/daemon.log"}},
        {"name": "search/web", "arguments": {"query": "sqlite wal mode single writer multiple readers"}},
        {"name": "db/query", "arguments": {"sql": "SELECT * FROM users WHERE SYNTAX_ERROR"}},
        
        # Silent Loop Anomaly (Consecutive identical failing calls)
        {"name": "auth/verify_session", "arguments": {"token": "expired_token_xyz"}},
        {"name": "auth/verify_session", "arguments": {"token": "expired_token_xyz"}},
        {"name": "auth/verify_session", "arguments": {"token": "expired_token_xyz"}},
        
        # Normal traffic
        {"name": "search/web", "arguments": {"query": "golang transparent stdio proxy latency benchmark"}},
        {"name": "db/query", "arguments": {"sql": "SELECT id, latency_ms FROM tool_traces ORDER BY created_at DESC LIMIT 10"}},
        {"name": "analytics/aggregate", "arguments": {"metric": "requests_per_second", "heavy": False}},
        
        # Slow Tool Anomaly (>5000ms delay)
        {"name": "analytics/aggregate", "arguments": {"metric": "yearly_event_lake_rollup", "heavy": True}},
        
        # Wrap up normal calls
        {"name": "db/query", "arguments": {"sql": "SELECT * FROM active_workers WHERE status = 'idle'"}},
        {"name": "fs/read_file", "arguments": {"path": "/Users/jang/Products/Product Hunt/README.md"}},
        {"name": "search/web", "arguments": {"query": "product hunt launch checklist 2026"}}
    ]

    print(f"3. Executing {len(calls)} diverse tool calls...")
    for idx, c in enumerate(calls, start=3):
        print(f"   [{idx}/{len(calls)+2}] Calling tool: {c['name']}...")
        send_rpc({
            "jsonrpc": "2.0",
            "id": idx,
            "method": "tools/call",
            "params": c
        })
        time.sleep(0.1)

    print("✅ Seed session completed successfully!")
    proc.terminate()
    proc.wait()

if __name__ == "__main__":
    main()
