#!/bin/bash

# Test script for SQL Schema Lint MCP Server SSE endpoint

echo "Testing SQL Schema Lint MCP Server SSE endpoint..."
echo "================================================="

# Server URL
SERVER_URL="http://localhost:8080"

echo -e "\nConnecting to SSE endpoint at $SERVER_URL/sse"
echo "Press Ctrl+C to stop..."
echo -e "\nSSE Events:\n"

# Connect to SSE endpoint and display events
curl -N -H "Accept: text/event-stream" $SERVER_URL/sse | while read -r line
do
    if [[ $line == "event:"* ]]; then
        echo -e "\n$line"
    elif [[ $line == "data:"* ]]; then
        echo "$line"
        # Try to pretty print JSON data
        data="${line#data: }"
        if command -v jq &> /dev/null; then
            echo "$data" | jq . 2>/dev/null || echo "$data"
        fi
    fi
done