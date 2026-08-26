# ⚡️ MCPOp — Open-Source Observability & Silent Failure Catcher for MCP Tools

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Zero--CGO-Pure%20Go%20SQLite-blue.svg" alt="Zero CGO">
  <img src="https://img.shields.io/badge/Latency-%3C1ms-brightgreen.svg" alt="Latency">
</p>

> **MCPOp** is a lightweight, zero-overhead **Transparent Stdio/SSE Proxy & Local Observability Dashboard** for AI Agents and Model Context Protocol (MCP) servers (Claude Desktop, Cursor, Antigravity).

---

## 🚨 The Problem: AI Agents Burning Tokens in Silence

When building or using **MCP Servers** with AI clients like Claude Desktop or Cursor:

1. **The Death Loop:** When an AI Agent calls a broken tool, it often retries blindly with the same failed payload 5–10 times ➡️ **Thousands of tokens burned in seconds.**
2. **Schema Mismatch & Hallucinations:** Agents hallucinate invalid arguments or missing required fields, causing obscure internal errors.
3. **Latency Spikes:** Sluggish tools (>5s) cause unexpected agent timeouts.
4. **Painful Debugging:** Testing a single failed tool call requires restarting Claude or writing manual stdin test scripts.

---

## 💡 The Solution: MCPOp

**MCPOp** acts as a micro-observability layer standing between your AI Client and any MCP Server (Python, Node.js, Go, Rust):

```
┌────────────────────────────────────────────────────────┐
│ AI Client (Claude Desktop / Cursor / Antigravity)      │
└──────────────────────────┬─────────────────────────────┘
                           │ Stdio / SSE JSON-RPC
┌──────────────────────────▼─────────────────────────────┐
│ MCPOp Engine (Single Standalone Go Binary)             │
│ ├── 1. Transparent Stdio Interceptor (<1ms overhead)  │
│ ├── 2. Heuristic Failure Engine (Goroutine Worker)     │
│ │    ├── Loop Detector (Consecutive failure alerts)    │
│ │    ├── Schema Mismatch / Hallucination Validator     │
│ │    └── Latency / Slow Tool Tracker                   │
│ ├── 3. Pure Go SQLite (modernc.org/sqlite - Zero CGO)  │
│ └── 4. Embedded Web Dashboard (http://localhost:4040)  │
└──────────────────────────┬─────────────────────────────┘
                           │ Child Process Stdio
┌──────────────────────────▼─────────────────────────────┐
│ Target MCP Server (Python, Node.js, Go, Rust, etc.)    │
└────────────────────────────────────────────────────────┘
```

---

## ✨ Features

- **🔍 Zero-Config Drop-in (<1ms Overhead):** Simply wrap your MCP server command with `mcpop run --`. No code changes required!
- **🛡️ Silent Failure Catcher:**
  - **Looping Alert:** Detects when an agent repeats failed calls ($\ge 2$ consecutive times).
  - **Schema Mismatch:** Catches missing required parameters or type mismatches against `tools/list` schemas.
  - **Slow Tool Warning:** Flags tools with execution times exceeding the configured threshold (default $>5000$ms).
- **📊 Realtime Live Waterfall Dashboard (`http://localhost:4040`):** Server-Sent Events (SSE) stream tool calls and anomaly alerts directly to a sleek dark-mode UI.
- **⚡ 1-Click Replay:** Inspect any tool call, edit JSON arguments, and re-execute directly from the browser to test tools in isolation.
- **📦 Single Standalone Binary:** Pure Go with embedded static assets (`//go:embed`) and pure Go SQLite. No Node.js or Python runtime required to run the proxy!

---

## 🚀 Quick Start

### 1. Installation

#### Using Go Install:
```bash
go install github.com/jangtrinh/mcpop/cmd/mcpop@latest
```

#### Or Build from Source:
```bash
git clone https://github.com/jangtrinh/mcpop.git
cd mcpop
make build
# Binary will be at ./bin/mcpop
```

---

### 2. Usage with Claude Desktop / Cursor

Update your `claude_desktop_config.json` or Cursor MCP settings:

#### Before:
```json
{
  "mcpServers": {
    "sqlite-server": {
      "command": "python",
      "args": ["/path/to/server.py"]
    }
  }
}
```

#### After (Observed with MCPOp):
```json
{
  "mcpServers": {
    "sqlite-server": {
      "command": "mcpop",
      "args": ["run", "--", "python", "/path/to/server.py"]
    }
  }
}
```

Then open **`http://localhost:4040`** to view the live dashboard!

---

### 3. CLI Commands & Flags

```bash
# Run with custom dashboard port:
mcpop run --port 8080 -- python server.py

# Set custom slow tool latency threshold (in ms):
mcpop run --slow-threshold 2000 -- node index.js

# Run in headless mode (no web UI, background SQLite logging only):
mcpop run --no-ui -- python server.py

# Open standalone web dashboard to view past traces and failures:
mcpop dashboard --port 4040
```

---

## 🧪 Testing

```bash
# Run unit & race tests
make test
# or: go test -count=1 -race ./...
```

---

## 📄 License

MIT License © 2026 Jang Trinh
