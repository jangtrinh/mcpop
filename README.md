# ⚡️ MCPOp — Open-Source Observability & Silent Failure Catcher for MCP Tools

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  <img src="https://img.shields.io/badge/Zero--CGO-Pure%20Go%20SQLite-blue.svg" alt="Zero CGO">
  <img src="https://img.shields.io/badge/Latency-%3C0.5ms-brightgreen.svg" alt="Latency">
</p>

> **MCPOp** is a lightweight, zero-overhead **Transparent Stdio/SSE Proxy & Local Observability Dashboard** for AI Agents and Model Context Protocol (MCP) servers (Claude Desktop, Cursor, Antigravity).

---

## 📸 Dashboard Showcase

![MCPOp Live Observability Dashboard](docs/assets/dashboard.png)

*Real-time trace waterfall, p50/p90/p99 telemetry analytics, heuristic silent failure alerts (looping, schema mismatches, slow tools), and 1-click isolated tool replay.*

---

## 🚨 The Problem: AI Agents Burning Tokens in Silence

When building or running **MCP Servers** with AI clients like Claude Desktop, Cursor, or autonomous agents:

1. **The Death Loop:** When an AI agent calls a broken tool, it often retries blindly with the same failed payload 5–10 times ➡️ **Thousands of tokens burned in seconds.**
2. **Schema Mismatch & Hallucinations:** Agents hallucinate invalid arguments or miss required properties, causing cryptic internal crashes.
3. **Latency Spikes:** Sluggish tools (>5s) cause unexpected agent timeouts and broken multi-turn workflows.
4. **Painful Debugging:** Testing a single failed tool call requires restarting Claude Desktop or writing manual stdin test scripts.

---

## 💡 The Solution: MCPOp

**MCPOp** acts as a transparent micro-observability layer standing between your AI Client and any MCP Server (Python, Node.js, Go, Rust):

```
┌────────────────────────────────────────────────────────┐
│ AI Client (Claude Desktop / Cursor / Antigravity)      │
└──────────────────────────┬─────────────────────────────┘
                           │ Stdio / SSE JSON-RPC
┌──────────────────────────▼─────────────────────────────┐
│ MCPOp Engine (Single Standalone Go Binary)             │
│ ├── 1. Transparent Stdio Interceptor (<0.5ms overhead)│
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

## ✨ Key Features

- **🔍 Zero-Config Drop-in (<0.5ms Overhead):** Simply prefix your MCP server command with `mcpop run --`. No code modifications required!
- **🛡️ Silent Failure Catcher:**
  - **Looping Alert:** Detects when an agent repeats failed calls ($\ge 2$ consecutive times with identical arguments).
  - **Schema Mismatch:** Validates arguments against `tools/list` schemas in real-time to catch missing required parameters or type violations.
  - **Slow Tool Warning:** Flags tools with execution times exceeding the configured threshold (default $>5000$ms).
- **📊 Realtime Live Waterfall Dashboard (`http://localhost:4040`):** Server-Sent Events (SSE) stream tool calls and anomaly alerts directly to a high-density, neutral B&W dashboard.
- **⚡ 1-Click Replay:** Inspect any tool call, edit JSON arguments, and re-execute directly in isolation from the browser without affecting client sessions.
- **📦 Single Standalone Binary:** Pure Go with embedded static assets (`//go:embed`) and pure Go SQLite. Zero external dependencies!

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
# Binary is at ./bin/mcpop
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

## 🧪 Testing & Mocking

```bash
# Run full unit & race tests:
make test

# Run realistic mock traffic generator:
python3 test/seed_data.py
```

---

## 📄 License

MIT License © 2026 Jang Trinh
