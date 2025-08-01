#!/bin/bash

echo "Testing server error handling..."

# Test without API key
echo "1. Testing without API key:"
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0.0"}},"id":1}' | ./bin/schema-lint-server 2>&1 | grep -v "Starting MCP"

# Test with dummy API key
echo -e "\n2. Testing with dummy API key:"
export GOOGLE_API_KEY="dummy_test_key"
cat << 'EOF' | ./bin/schema-lint-server 2>&1 | jq -r 'select(.id==2) | .error // .result' | head -20
{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"1.0.0"}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"lint_schema","arguments":{"schema_path":"./examples/schemas/database.sql","rules_path":"./examples/rules/sql-schema-rules.md","provider":"gemini","dialect":"postgres","output_format":"text"}},"id":2}
EOF

echo -e "\nDone!"