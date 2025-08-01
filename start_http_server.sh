#!/bin/bash

# Script to start the MCP server with HTTP transport

echo "Starting SQL Schema Lint MCP Server with HTTP transport..."
echo "========================================================="

# Set environment variables
export MCP_TRANSPORT=http
export PORT=8070

# Check if API key is set
if [ -z "$GOOGLE_API_KEY" ] && [ -z "$ANTHROPIC_API_KEY" ] && [ -z "$OPENAI_API_KEY" ]; then
    echo "Warning: No AI provider API key found in environment."
    echo "Please set one of: GOOGLE_API_KEY, ANTHROPIC_API_KEY, or OPENAI_API_KEY"
    echo ""
fi

echo "Configuration:"
echo "  Transport: HTTP"
echo "  Port: $PORT"
echo "  Endpoints:"
echo "    - Root: http://localhost:$PORT/"
echo "    - Health: http://localhost:$PORT/healthz"
echo "    - SSE: http://localhost:$PORT/sse"
echo "    - Message: http://localhost:$PORT/message"
echo ""

# Run the server
go run cmd/server/main.go