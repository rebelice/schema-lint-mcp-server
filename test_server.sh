#!/bin/bash

# Test script for SQL Schema Lint MCP Server

echo "Testing SQL Schema Lint MCP Server..."
echo "===================================="

# Server URL
SERVER_URL="http://localhost:8070"

# Test 1: Root endpoint
echo -e "\n1. Testing root endpoint..."
curl -s -X GET $SERVER_URL/ | jq .

# Test 2: Health endpoint
echo -e "\n2. Testing health endpoint..."
curl -s -X GET $SERVER_URL/healthz | jq .

# Test 3: Initialize session
echo -e "\n3. Testing initialize..."
INIT_RESPONSE=$(curl -s -X POST $SERVER_URL/message \
  -H "Content-Type: application/json" \
  -d '{
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
  }')
echo $INIT_RESPONSE | jq .

# Test 4: List tools
echo -e "\n4. Testing tools/list..."
TOOLS_RESPONSE=$(curl -s -X POST $SERVER_URL/message \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "id": 2
  }')
echo $TOOLS_RESPONSE | jq .

# Test 5: Call lint_schema tool
echo -e "\n5. Testing lint_schema tool..."
LINT_RESPONSE=$(curl -s -X POST $SERVER_URL/message \
  -H "Content-Type: application/json" \
  -d '{
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
  }')
echo $LINT_RESPONSE | jq .

echo -e "\nTest completed!"