#!/usr/bin/env python3
"""
Rich Mock MCP Server for MCPOp End-to-End Observability Testing
Supports real-world tool specifications and edge cases.
"""
import sys
import json
import time

def main():
    while True:
        line = sys.stdin.readline()
        if not line:
            break
        line = line.strip()
        if not line:
            continue

        try:
            req = json.loads(line)
        except json.JSONDecodeError:
            continue

        method = req.get("method")
        msg_id = req.get("id")

        if method == "initialize":
            resp = {
                "jsonrpc": "2.0",
                "id": msg_id,
                "result": {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {"tools": {}},
                    "serverInfo": {
                        "name": "enterprise-mock-mcp",
                        "version": "2.1.0"
                    }
                }
            }
            sys.stdout.write(json.dumps(resp) + "\n")
            sys.stdout.flush()

        elif method == "tools/list":
            resp = {
                "jsonrpc": "2.0",
                "id": msg_id,
                "result": {
                    "tools": [
                        {
                            "name": "db/query",
                            "description": "Execute read-only SQL queries on PostgreSQL cluster",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "sql": {"type": "string"},
                                    "limit": {"type": "integer"}
                                },
                                "required": ["sql"]
                            }
                        },
                        {
                            "name": "search/web",
                            "description": "Perform real-time search across knowledge sources",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "query": {"type": "string"},
                                    "max_results": {"type": "integer"}
                                },
                                "required": ["query"]
                            }
                        },
                        {
                            "name": "fs/read_file",
                            "description": "Read contents of a file on disk",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "path": {"type": "string"}
                                },
                                "required": ["path"]
                            }
                        },
                        {
                            "name": "ai/embed_text",
                            "description": "Generate vector embeddings for given text",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "text": {"type": "string"},
                                    "model": {"type": "string"}
                                },
                                "required": ["text", "model"]
                            }
                        },
                        {
                            "name": "auth/verify_session",
                            "description": "Validate API token and session claims",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "token": {"type": "string"}
                                },
                                "required": ["token"]
                            }
                        },
                        {
                            "name": "analytics/aggregate",
                            "description": "Compute heavy aggregations across event lake",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "metric": {"type": "string"},
                                    "heavy": {"type": "boolean"}
                                }
                            }
                        }
                    ]
                }
            }
            sys.stdout.write(json.dumps(resp) + "\n")
            sys.stdout.flush()

        elif method == "tools/call":
            params = req.get("params", {})
            tool_name = params.get("name")
            args = params.get("arguments", {})

            if tool_name == "db/query":
                sql = args.get("sql", "")
                if "SYNTAX_ERROR" in sql:
                    resp = {
                        "jsonrpc": "2.0",
                        "id": msg_id,
                        "result": {
                            "content": [{"type": "text", "text": "PostgreSQL Error: syntax error at or near 'SYNTAX_ERROR'"}],
                            "isError": True
                        }
                    }
                else:
                    time.sleep(0.04)
                    resp = {
                        "jsonrpc": "2.0",
                        "id": msg_id,
                        "result": {
                            "content": [{"type": "text", "text": json.dumps([{"id": 101, "name": "Antigravity Pro", "status": "active"}, {"id": 102, "name": "AgentKit Cloud", "status": "active"}])}],
                            "isError": False
                        }
                    }

            elif tool_name == "search/web":
                query = args.get("query", "")
                time.sleep(0.12)
                resp = {
                    "jsonrpc": "2.0",
                    "id": msg_id,
                    "result": {
                        "content": [{"type": "text", "text": f"Found 12 matching articles for query '{query}'. Best match: MCPOp - Zero-overhead MCP proxy."}],
                        "isError": False
                    }
                }

            elif tool_name == "fs/read_file":
                path = args.get("path", "")
                time.sleep(0.02)
                resp = {
                    "jsonrpc": "2.0",
                    "id": msg_id,
                    "result": {
                        "content": [{"type": "text", "text": f"# Config for {path}\nport = 4040\nlog_level = 'info'\n"}],
                        "isError": False
                    }
                }

            elif tool_name == "ai/embed_text":
                text = args.get("text", "")
                time.sleep(0.08)
                resp = {
                    "jsonrpc": "2.0",
                    "id": msg_id,
                    "result": {
                        "content": [{"type": "text", "text": json.dumps({"dimensions": 1536, "norm": 0.9998, "sample": [0.012, -0.045, 0.089]})}],
                        "isError": False
                    }
                }

            elif tool_name == "auth/verify_session":
                token = args.get("token", "")
                if token == "expired_token_xyz":
                    time.sleep(0.03)
                    resp = {
                        "jsonrpc": "2.0",
                        "id": msg_id,
                        "result": {
                            "content": [{"type": "text", "text": "HTTP 401 Unauthorized: token expired at 2026-08-25T12:00:00Z"}],
                            "isError": True
                        }
                    }
                else:
                    time.sleep(0.02)
                    resp = {
                        "jsonrpc": "2.0",
                        "id": msg_id,
                        "result": {
                            "content": [{"type": "text", "text": json.dumps({"user_id": "usr_9981", "role": "admin", "valid": True})}],
                            "isError": False
                        }
                    }

            elif tool_name == "analytics/aggregate":
                is_heavy = args.get("heavy", False)
                if is_heavy:
                    time.sleep(5.2)  # Triggers slow tool heuristic (>5000ms)
                else:
                    time.sleep(0.45)
                resp = {
                    "jsonrpc": "2.0",
                    "id": msg_id,
                    "result": {
                        "content": [{"type": "text", "text": json.dumps({"p50_latency": "18ms", "total_requests": 142000, "error_rate": "0.02%"})}],
                        "isError": False
                    }
                }
            else:
                resp = {
                    "jsonrpc": "2.0",
                    "id": msg_id,
                    "error": {
                        "code": -32601,
                        "message": f"Tool not found: {tool_name}"
                    }
                }

            sys.stdout.write(json.dumps(resp) + "\n")
            sys.stdout.flush()

if __name__ == "__main__":
    main()
