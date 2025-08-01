#!/bin/bash

# Test script for SQL Schema Lint MCP Server (stdio transport)

echo "Testing SQL Schema Lint MCP Server (stdio transport)..."
echo "====================================================="

# Test initialize
echo -e "\n1. Testing initialize..."
echo '{
    "jsonrpc": "2.0",
    "method": "initialize",
    "params": {
        "protocolVersion": "2024-11-05",
        "clientInfo": {
            "name": "test-client",
            "version": "1.0.0"
        }
    },
    "id": 1
}' | go run cmd/server/main.go

# Test list tools
echo -e "\n\n2. Testing tools/list..."
echo '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "id": 2
}' | go run cmd/server/main.go

# Test lint_schema tool
echo -e "\n\n3. Testing lint_schema tool..."
echo '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
        "name": "lint_schema",
        "arguments": {
            "schema_path": "./examples/schemas/database.sql",
            "rules_path": "./examples/rules/sql-schema-rules.md",
            "provider": "gemini",
            "dialect": "postgres",
            "output_format": "json"
        }
    },
    "id": 3
}' | go run cmd/server/main.go

echo -e "\n\nTest completed!"