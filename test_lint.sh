#!/bin/bash

echo "Running schema lint on examples..."

# Create a temporary file with both messages
cat > /tmp/lint_messages.jsonl << 'EOF'
{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test-client","version":"1.0.0"}},"id":1}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"lint_schema","arguments":{"schema_path":"./examples/schemas/database.sql","rules_path":"./examples/rules/sql-schema-rules.md","provider":"gemini","dialect":"postgres","output_format":"markdown"}},"id":2}
EOF

# Run the server with the messages
export GOOGLE_API_KEY="${GOOGLE_API_KEY:-test_key}"
cat /tmp/lint_messages.jsonl | go run cmd/server/main.go 2>&1 | jq -r 'select(.id==2) | .result.content[0].text // .error.message // .'

rm /tmp/lint_messages.jsonl