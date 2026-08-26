#!/usr/bin/env python3
"""
Mock MCP Server for testing MCPOp proxy
Supports:
- initialize
- tools/list
- tools/call (calculate, fail_tool, slow_tool)
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
                    "capabilities": {
                        "tools": {}
                    },
                    "serverInfo": {
                        "name": "mock-mcp-server",
                        "version": "1.0.0"
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
                            "name": "calculate",
                            "description": "Perform basic arithmetic",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "expr": {"type": "string"}
                                },
                                "required": ["expr"]
                            }
                        },
                        {
                            "name": "fail_tool",
                            "description": "Tool that always fails",
                            "inputSchema": {"type": "object"}
                        },
                        {
                            "name": "slow_tool",
                            "description": "Tool that takes time to respond",
                            "inputSchema": {"type": "object"}
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

            if tool_name == "calculate":
                expr = args.get("expr", "0")
                try:
                    # Safe eval for simple math
                    val = eval(expr, {"__builtins__": None}, {})
                    result_content = [{"type": "text", "text": f"Result: {val}"}]
                    is_error = False
                except Exception as e:
                    result_content = [{"type": "text", "text": f"Error: {e}"}]
                    is_error = True

                resp = {
                    "jsonrpc": "2.0",
                    "id": msg_id,
                    "result": {
                        "content": result_content,
                        "isError": is_error
                    }
                }
            elif tool_name == "fail_tool":
                resp = {
                    "jsonrpc": "2.0",
                    "id": msg_id,
                    "result": {
                        "content": [{"type": "text", "text": "Database connection timeout (simulated)"}],
                        "isError": True
                    }
                }
            elif tool_name == "slow_tool":
                delay = float(args.get("delay", 1.0))
                time.sleep(delay)
                resp = {
                    "jsonrpc": "2.0",
                    "id": msg_id,
                    "result": {
                        "content": [{"type": "text", "text": f"Finished after {delay}s"}],
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
