# MCPOp

Transparent observability proxy and silent failure catcher for Model Context Protocol (MCP) servers.

![MCPOp Live Observability Dashboard](docs/assets/dashboard.png)

---

## Why this exists

If you use Claude Desktop, Cursor, or autonomous agents with MCP servers, you have likely run into this:

You install a bunch of MCP tools, start a coding session, and suddenly your token budget evaporates while the agent produces garbage output.

Here is what is actually happening under the hood:

1. **The Retry Death Loop:** When an MCP tool fails silently (expired token, database timeout, internal 500), the agent does not just stop. It aggressively retries the exact same broken payload 5 to 10 times in a row. Because every retry re-transmits the entire conversation context window, you burn 50,000+ tokens in 30 seconds for zero result.
2. **Schema Mismatches:** The agent hallucinates arguments or drops a required property. The tool crashes, but because everything runs over background stdio pipes, you see zero logs.
3. **Latency Black Holes:** A sluggish tool (>5s) stalls the agent workflow until the client times out.
4. **Painful Debug Loops:** Testing a single broken tool call used to require restarting Claude Desktop or writing manual stdin test harnesses.

MCPOp sits invisibly between your AI client and your MCP server to catch these failures in real time before they burn your wallet.

---

## What MCPOp does

MCPOp is a single, zero-dependency Go binary with under 0.5ms proxy overhead.

- **Loop Detection:** Flags consecutive failed tool calls with identical arguments before they spiral into token-burning loops.
- **Real-Time Schema Validation:** Cross-checks argument payloads against the server's `tools/list` schema to immediately pinpoint missing required parameters.
- **Latency Tracking:** Measures execution time per call and flags slow operations exceeding threshold.
- **1-Click Replay:** Includes an embedded, minimalist B&W web dashboard with live trace waterfalls. Inspect any past payload, edit JSON arguments inline, and re-execute directly in isolation from your browser without restarting your AI client.
- **Zero Runtime Dependencies:** Packaged as a single standalone binary with embedded SQLite and Web UI. No Node.js, Python, or CGO compiler required.

---

## Architecture

```
┌────────────────────────────────────────────────────────┐
│ AI Client (Claude Desktop / Cursor / Antigravity)      │
└──────────────────────────┬─────────────────────────────┘
                           │ Stdio / SSE JSON-RPC
┌──────────────────────────▼─────────────────────────────┐
│ MCPOp Core (<0.5ms Proxy Overhead)                     │
│ ├── Transparent Stdio Pipe Interceptor                 │
│ ├── Async Heuristic Failure Worker                     │
│ │    ├── Retry Loop Catcher (consecutive failures)     │
│ │    ├── Schema Mismatch & Property Validator          │
│ │    └── Latency & Slow Tool Profiler                  │
│ ├── Pure Go SQLite Persistence (~/.mcpop/data.db)      │
│ └── Embedded Web Dashboard (http://localhost:4040)     │
└──────────────────────────┬─────────────────────────────┘
                           │ Child Subprocess Stdio
┌──────────────────────────▼─────────────────────────────┐
│ Target MCP Server (Python, Node.js, Go, Rust, etc.)    │
└────────────────────────────────────────────────────────┘
```

---

## Quick Start

### 1. Install

Using Go:
```bash
go install github.com/jangtrinh/mcpop/cmd/mcpop@latest
```

Or build from source:
```bash
git clone https://github.com/jangtrinh/mcpop.git
cd mcpop
make build
# Binary is at ./bin/mcpop
```

### 2. Configure with Claude Desktop or Cursor

Prefix your existing MCP server command with `mcpop run --`.

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

#### After:
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

Open `http://localhost:4040` to view the live dashboard.

---

## CLI Reference

```bash
# Wrap an MCP server command
mcpop run -- python server.py

# Custom dashboard port
mcpop run --port 8080 -- python server.py

# Custom slow tool threshold (in ms, default: 5000)
mcpop run --slow-threshold 2000 -- node index.js

# Headless mode (no web server, SQLite background logging only)
mcpop run --no-ui -- python server.py

# Open standalone dashboard to view past traces and failures
mcpop dashboard --port 4040

# Print version
mcpop version
```

---

## Development and Testing

```bash
# Run unit and race detection tests
make test

# Generate realistic mock traffic (20+ tool calls with anomalies)
python3 test/seed_data.py
```

---

## License

MIT License. Free and open source.
