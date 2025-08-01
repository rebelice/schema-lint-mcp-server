# SQL Schema Lint MCP Server

## Overview

This MCP (Model Context Protocol) server provides an AI-powered SQL schema validation tool that uses AI APIs (Gemini, Claude, GPT-4, etc.) to lint SQL schema files against user-defined rules. The server combines the intelligence of AI models with structured rule definitions to provide comprehensive schema analysis, best practice recommendations, and intelligent suggestions.

## Features

### `lint_schema` Tool

The primary tool exposed by this MCP server that validates SQL schema files against custom lint rules.

#### Parameters

- `schema_path` (string, required): Path to the SQL schema file to be validated (.sql, .ddl)
- `rules_path` (string, required): Path to the rules configuration file (.md)
- `provider` (string, optional): AI provider to use (gemini, claude, openai). Default: gemini
- `dialect` (string, optional): SQL dialect (postgres, mysql, sqlite, sqlserver, oracle). If not specified, auto-detects from content
- `output_format` (string, optional): Output format for lint results (json, text, markdown). Default: json

#### Example Usage

```json
{
  "tool": "lint_schema",
  "parameters": {
    "schema_path": "./schemas/database.sql",
    "rules_path": "./rules/sql-schema-rules.md",
    "provider": "gemini",
    "dialect": "postgres",
    "output_format": "json"
  }
}
```

## Rules Configuration

Rules are defined in Markdown format for better readability and documentation. Each rule is defined in a section with specific metadata.

### SQL Schema Rules Example (`sql-schema-rules.md`)

```markdown
# SQL Schema Lint Rules

## Rule: table-naming-convention
- **Severity**: error
- **Target**: tables.*
- **Check**: name_pattern("^[a-z][a-z0-9_]*$")

Table names must be lowercase with underscores (snake_case).

## Rule: primary-key-required
- **Severity**: error
- **Target**: tables.*
- **Check**: has_primary_key()

Every table must have a primary key defined.

## Rule: column-naming-convention
- **Severity**: warning
- **Target**: tables.*.columns.*
- **Check**: name_pattern("^[a-z][a-z0-9_]*$")

Column names should use snake_case for consistency.

## Rule: foreign-key-naming
- **Severity**: warning
- **Target**: tables.*.foreign_keys.*
- **Check**: name_pattern("^fk_[a-z]+_[a-z]+$")

Foreign key constraints should follow the pattern: fk_<table>_<referenced_table>.

## Rule: index-naming-convention
- **Severity**: warning
- **Target**: tables.*.indexes.*
- **Check**: name_pattern("^idx_[a-z]+_[a-z_]+$")

Indexes should follow the pattern: idx_<table>_<columns>.

## Rule: no-reserved-words
- **Severity**: error
- **Target**: tables.*, tables.*.columns.*
- **Check**: not_reserved_word()

Table and column names must not use SQL reserved words.

## Rule: timestamp-columns
- **Severity**: info
- **Target**: tables.*
- **Check**: has_columns(["created_at", "updated_at"])

Tables should have created_at and updated_at timestamp columns.
```

### Advanced SQL Rules Example (`advanced-sql-rules.md`)

```markdown
# Advanced SQL Schema Lint Rules

## Rule: varchar-length-specified
- **Severity**: error
- **Target**: tables.*.columns[type=varchar]
- **Check**: has_length_constraint()

VARCHAR columns must specify a maximum length.

## Rule: enum-check-constraint
- **Severity**: warning
- **Target**: tables.*.columns[has_check_constraint]
- **Check**: check_constraint_values_documented()

Columns with CHECK constraints for enum-like values should be documented.

## Rule: cascade-delete-restriction
- **Severity**: warning
- **Target**: tables.*.foreign_keys[on_delete=CASCADE]
- **Check**: has_comment()

Foreign keys with CASCADE DELETE must have a comment explaining the impact.

## Rule: unique-constraint-naming
- **Severity**: warning
- **Target**: tables.*.constraints[type=unique]
- **Check**: name_pattern("^uq_[a-z]+_[a-z_]+$")

Unique constraints should follow the pattern: uq_<table>_<columns>.

## Rule: partition-key-index
- **Severity**: error
- **Target**: tables[partitioned=true]
- **Check**: partition_key_indexed()

Partitioned tables must have an index on the partition key.
```

### Rules Markdown Format

Each rule in the Markdown file should follow this structure:

```markdown
## Rule: [rule-id]
- **Severity**: error|warning|info
- **Target**: [JSONPath-like expression]
- **Check**: [check-expression]
- **Tags**: [optional, comma, separated, tags]

[Rule description explaining why this rule exists and what it checks]

### Good Example (optional)
```sql
-- Example of correct usage
```

### Bad Example (optional)
```sql
-- Example of what violates this rule
```
```

## Supported SQL Dialects

- **PostgreSQL** (9.5+)
- **MySQL** (5.7+, 8.0+)
- **SQLite** (3.x)
- **Microsoft SQL Server** (2016+)
- **Oracle** (12c+)
- **MariaDB** (10.x)
- **Amazon Redshift**
- **Google BigQuery**
- **Snowflake**

## Output Formats

### JSON Output
```json
{
  "schema_file": "./schemas/database.sql",
  "dialect": "postgres",
  "total_issues": 3,
  "errors": 1,
  "warnings": 2,
  "issues": [
    {
      "rule_id": "primary-key-required",
      "severity": "error",
      "object_type": "table",
      "object_name": "users",
      "message": "Every table must have a primary key defined",
      "line": 15,
      "column": 1
    },
    {
      "rule_id": "column-naming-convention",
      "severity": "warning",
      "object_type": "column",
      "object_name": "users.firstName",
      "message": "Column names should use snake_case for consistency",
      "line": 17,
      "column": 3,
      "suggestion": "first_name"
    }
  ]
}
```

### Text Output
```
SQL Schema Lint Results for ./schemas/database.sql
==================================================
Dialect: postgres
Total Issues: 3 (1 error, 2 warnings)

ERROR: Table 'users' (line 15)
  Rule: primary-key-required
  Every table must have a primary key defined

WARNING: Column 'users.firstName' (line 17)
  Rule: column-naming-convention
  Column names should use snake_case for consistency
  Suggestion: first_name

WARNING: Foreign Key 'users_orders_fk' (line 45)
  Rule: foreign-key-naming
  Foreign key constraints should follow the pattern: fk_<table>_<referenced_table>
  Suggestion: fk_users_orders
```

## AI-Powered Analysis

The server leverages AI APIs to provide intelligent schema analysis:

### How It Works

1. **Schema Parsing**: The SQL schema file is parsed to extract structure
2. **Rule Compilation**: Markdown rules are converted into AI prompts
3. **AI Analysis**: The AI model analyzes the schema against each rule
4. **Result Aggregation**: AI responses are structured into actionable feedback

### AI Prompt Template

For each rule, the server generates a prompt like:

```
Analyze this SQL schema for the following rule:
Rule: {rule_name}
Severity: {severity}
Description: {description}

SQL Schema:
{schema_content}

Check if the schema violates this rule. For each violation found, provide:
1. The specific object (table/column/constraint) that violates the rule
2. Line number where the violation occurs
3. A clear explanation of why it violates the rule
4. A suggested fix

Format your response as JSON.
```

## Implementation Details

The MCP server should:

1. **Parse SQL DDL Statements**: Extract schema structure for AI analysis
2. **AI Integration**: 
   - Support multiple AI providers (Gemini API, Claude API, OpenAI API)
   - Handle API rate limiting and retries
   - Cache AI responses for efficiency
3. **Rule Processing**: Convert Markdown rules into effective AI prompts
4. **Intelligent Analysis**:
   - Beyond pattern matching - understand context and intent
   - Provide intelligent suggestions based on best practices
   - Detect subtle issues that regex-based linters might miss
5. **Result Processing**: Parse AI responses and structure them consistently
6. **Performance Optimization**: 
   - Batch multiple rules into single AI calls when possible
   - Implement response caching
7. **Fallback Mechanisms**: Basic rule checking when AI is unavailable

## MCP HTTP Transport Protocol

### Transport Overview

The server implements the MCP HTTP transport protocol with Server-Sent Events (SSE) support for streaming capabilities. This follows the MCP specification version 2024-11-05 and later.

### HTTP Endpoints

#### 1. Root Endpoint (`GET /`)
Returns server information and available endpoints:

```json
{
  "name": "sql-schema-lint",
  "version": "1.0.0",
  "description": "AI-powered MCP server for SQL schema linting",
  "endpoints": {
    "sse": "/sse",
    "message": "/message"
  },
  "transport": "http"
}
```

#### 2. SSE Endpoint (`GET /sse`)
Establishes a Server-Sent Events stream for server-to-client communication:
- Content-Type: `text/event-stream`
- Used for asynchronous notifications and streaming responses
- Supports connection resumption with `Last-Event-ID` header

#### 3. Message Endpoint (`POST /message`)
Handles JSON-RPC 2.0 requests for client-to-server communication:
- Content-Type: `application/json`
- Accept: `application/json, text/event-stream`
- Body: Single JSON-RPC request or batch array

#### 4. Health Check (`GET /healthz`)
Returns server health status:

```json
{
  "status": "healthy",
  "service": "sql-schema-lint",
  "version": "1.0.0",
  "uptime": "running"
}
```

### JSON-RPC Methods

#### 1. Initialize (`initialize`)
Establishes a new MCP session:

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "clientInfo": {
      "name": "claude-code",
      "version": "1.0.0"
    }
  },
  "id": 1
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "sql-schema-lint",
      "version": "1.0.0"
    }
  },
  "id": 1
}
```

#### 2. List Tools (`tools/list`)
Returns available tools:

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/list",
  "id": 2
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "tools": [
      {
        "name": "lint_schema",
        "description": "Validate SQL schema files against user-defined rules using AI",
        "inputSchema": {
          "type": "object",
          "properties": {
            "schema_path": {
              "type": "string",
              "description": "Path to the SQL schema file to be validated (.sql, .ddl)"
            },
            "rules_path": {
              "type": "string",
              "description": "Path to the rules configuration file (.md)"
            },
            "provider": {
              "type": "string",
              "description": "AI provider to use (gemini, claude, openai)",
              "enum": ["gemini", "claude", "openai"]
            },
            "dialect": {
              "type": "string",
              "description": "SQL dialect (postgres, mysql, sqlite, sqlserver, oracle)",
              "enum": ["postgres", "mysql", "sqlite", "sqlserver", "oracle"]
            },
            "output_format": {
              "type": "string",
              "description": "Output format for lint results",
              "enum": ["json", "text", "markdown"]
            }
          },
          "required": ["schema_path", "rules_path"]
        }
      }
    ]
  },
  "id": 2
}
```

#### 3. Call Tool (`tools/call`)
Executes a tool with provided arguments:

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "lint_schema",
    "arguments": {
      "schema_path": "./schemas/database.sql",
      "rules_path": "./rules/sql-schema-rules.md",
      "provider": "gemini",
      "dialect": "postgres",
      "output_format": "json"
    }
  },
  "id": 3
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"schema_file\": \"./schemas/database.sql\", \"dialect\": \"postgres\", \"total_issues\": 3, ...}"
      }
    ]
  },
  "id": 3
}
```

### HTTP Headers

#### Required Headers
- **Content-Type**: `application/json` for requests
- **Accept**: `application/json, text/event-stream` for supporting both response types

#### CORS Headers
The server includes CORS headers for browser-based clients:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Accept, Last-Event-ID`

### Session Management

1. **Session Initialization**: Client calls `initialize` method to establish session
2. **Session ID**: Server may assign session ID for stateful operations
3. **Session Persistence**: Sessions persist across requests using session tokens

### Error Handling

JSON-RPC error responses follow standard error codes:

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32700,    // Parse error
    "message": "Parse error",
    "data": "Additional error details"
  },
  "id": null
}
```

Standard error codes:
- `-32700`: Parse error
- `-32600`: Invalid request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

### SSE Message Format

Server-sent events follow this format:

```
event: message
data: {"jsonrpc": "2.0", "method": "notification", "params": {...}}

event: ping
data: {"timestamp": 1234567890}

event: error
data: {"code": -32603, "message": "Internal error"}
```

### Authentication (Optional)

For production deployments, implement OAuth 2.1 or API key authentication:

#### API Key Authentication
```
Authorization: Bearer <api-key>
```

#### OAuth 2.1 Flow
1. Client requests authorization
2. Server redirects to auth provider
3. Client receives access token
4. Client includes token in requests

### Rate Limiting

The server implements rate limiting to prevent abuse:
- Default: 30 requests per minute per client
- Configurable via environment variables
- Returns 429 Too Many Requests when exceeded

### Connection Management

1. **Keep-Alive**: HTTP/1.1 persistent connections supported
2. **Timeouts**: Configurable request timeout (default: 30s)
3. **SSE Reconnection**: Clients should implement exponential backoff
4. **Graceful Shutdown**: Server completes in-flight requests before closing

### Batch Requests

The server supports JSON-RPC batch requests:

```json
[
  {"jsonrpc": "2.0", "method": "tools/list", "id": 1},
  {"jsonrpc": "2.0", "method": "tools/call", "params": {...}, "id": 2}
]
```

### Protocol Versioning

- Current protocol version: `2024-11-05`
- Version negotiation during initialization
- Backward compatibility for older versions where possible

## Configuration

The server can be configured via environment variables or a config file:

```yaml
server:
  name: "sql-schema-lint"
  version: "1.0.0"
  description: "AI-powered MCP server for SQL schema linting"

ai:
  default_provider: "gemini"
  providers:
    gemini:
      api_key: "${GOOGLE_API_KEY}"
      model: "gemini-2.0-flash-exp"
      max_tokens: 8192
    claude:
      api_key: "${ANTHROPIC_API_KEY}"
      model: "claude-3-sonnet-20240229"
      max_tokens: 4096
    openai:
      api_key: "${OPENAI_API_KEY}"
      model: "gpt-4-turbo-preview"
      max_tokens: 4096
  
  cache:
    enabled: true
    ttl_seconds: 3600
    
  rate_limiting:
    max_requests_per_minute: 30
    retry_attempts: 3
    retry_delay_ms: 1000

defaults:
  output_format: "json"
  severity_threshold: "warning"
  dialect: "postgres"
  provider: "gemini"
  
parsers:
  sql:
    case_sensitive: false
    allow_quoted_identifiers: true
    max_identifier_length: 63
  
  postgres:
    extensions: ["uuid-ossp", "postgis"]
    
  mysql:
    sql_mode: "TRADITIONAL"
    
  sqlite:
    foreign_keys: true
    
rules:
  batch_size: 5  # Number of rules to check in a single AI call
  timeout_seconds: 30
```

## Error Handling

The tool should handle various error scenarios:

- Invalid SQL file path
- SQL syntax errors
- Invalid rules configuration
- Unsupported SQL dialects
- Parser errors
- AI API errors (rate limits, timeouts, invalid responses)
- Missing API credentials
- Circular foreign key dependencies
- Missing referenced tables/columns

All errors should be returned in a structured format that MCP clients can understand and display appropriately to users.

## Advantages of AI-Powered Linting

1. **Contextual Understanding**: AI can understand the intent behind schema design, not just syntax
2. **Intelligent Suggestions**: Provides context-aware fixes rather than generic recommendations
3. **Natural Language Rules**: Rules can be written in plain English without complex regex patterns
4. **Adaptive Analysis**: Can identify issues based on industry best practices not explicitly defined in rules
5. **Explanation Quality**: Provides detailed explanations of why something is an issue
6. **Cross-Table Analysis**: Can understand complex relationships and dependencies across the entire schema
7. **Evolution**: As AI models improve, the linting becomes more sophisticated without code changes

## MCP HTTP Server Implementation Examples

### Basic HTTP Server Setup (Go)

```go
package main

import (
    "github.com/mark3labs/mcp-go/server"
    "github.com/rebeliceyang/schema-lint-mcp-server/internal/server"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8070"
    }
    
    // Create MCP server
    srv := server.NewMCPServer(config, aiConfig)
    
    // Create SSE server with proper configuration
    baseURL := fmt.Sprintf("http://localhost:%s", port)
    sseServer := server.NewSSEServer(srv.GetMCPServer(),
        server.WithBaseURL(baseURL),
        server.WithSSEEndpoint("/sse"),
        server.WithMessageEndpoint("/message"),
    )
    
    // Setup HTTP routes
    mux := http.NewServeMux()
    mux.HandleFunc("/", handleRoot)
    mux.HandleFunc("/healthz", handleHealth)
    mux.Handle("/sse", sseServer)
    mux.Handle("/message", sseServer)
    
    // Start server
    log.Printf("Starting MCP HTTP/SSE server on port %s", port)
    httpServer := &http.Server{
        Addr:    fmt.Sprintf(":%s", port),
        Handler: mux,
    }
    
    if err := httpServer.ListenAndServe(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

### Client Connection Example

```javascript
// JavaScript/TypeScript client example
class MCPClient {
    constructor(baseUrl) {
        this.baseUrl = baseUrl;
        this.sessionId = null;
    }
    
    async initialize() {
        const response = await fetch(`${this.baseUrl}/message`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json, text/event-stream'
            },
            body: JSON.stringify({
                jsonrpc: '2.0',
                method: 'initialize',
                params: {
                    protocolVersion: '2024-11-05',
                    clientInfo: {
                        name: 'mcp-client',
                        version: '1.0.0'
                    }
                },
                id: 1
            })
        });
        
        const result = await response.json();
        return result;
    }
    
    async callTool(toolName, arguments) {
        const response = await fetch(`${this.baseUrl}/message`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({
                jsonrpc: '2.0',
                method: 'tools/call',
                params: {
                    name: toolName,
                    arguments: arguments
                },
                id: Date.now()
            })
        });
        
        return await response.json();
    }
    
    connectSSE() {
        const eventSource = new EventSource(`${this.baseUrl}/sse`);
        
        eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            console.log('SSE message:', data);
        };
        
        eventSource.onerror = (error) => {
            console.error('SSE error:', error);
            // Implement reconnection logic
        };
        
        return eventSource;
    }
}
```

### Testing MCP HTTP Server

```bash
#!/bin/bash
# Test script for MCP HTTP server

# Test root endpoint
echo "Testing root endpoint..."
curl -X GET http://localhost:8070/

# Test health endpoint
echo -e "\n\nTesting health endpoint..."
curl -X GET http://localhost:8070/healthz

# Test initialize
echo -e "\n\nTesting initialize..."
curl -X POST http://localhost:8070/message \
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
  }'

# Test tools/list
echo -e "\n\nTesting tools/list..."
curl -X POST http://localhost:8070/message \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "id": 2
  }'

# Test tool call
echo -e "\n\nTesting tool call..."
curl -X POST http://localhost:8070/message \
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
  }'
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mcp-server cmd/server/main_mcp_http.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mcp-server .
EXPOSE 8070
CMD ["./mcp-server"]
```

### Security Best Practices

1. **TLS/HTTPS**: Always use HTTPS in production
```go
// Example with TLS
httpServer := &http.Server{
    Addr:      ":443",
    Handler:   mux,
    TLSConfig: tlsConfig,
}
httpServer.ListenAndServeTLS("cert.pem", "key.pem")
```

2. **Request Validation**: Validate all incoming requests
```go
func validateRequest(req json.RawMessage) error {
    // Check JSON structure
    // Validate method names
    // Verify parameter types
    // Sanitize file paths
}
```

3. **Authentication Middleware**:
```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !isValidToken(token) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Performance Considerations

1. **Connection Pooling**: Reuse HTTP connections
2. **Request Batching**: Support batch JSON-RPC requests
3. **Caching**: Cache AI responses with TTL
4. **Rate Limiting**: Implement per-client rate limits
5. **Async Processing**: Use goroutines for concurrent requests

### Monitoring and Observability

```go
// Prometheus metrics example
var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mcp_requests_total",
            Help: "Total number of MCP requests",
        },
        []string{"method", "status"},
    )
    
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "mcp_request_duration_seconds",
            Help: "Duration of MCP requests",
        },
        []string{"method"},
    )
)
```

### MCP Client Libraries

- **Go**: `github.com/mark3labs/mcp-go`
- **Python**: `mcp-python`
- **TypeScript**: `@modelcontextprotocol/sdk`
- **Rust**: `mcp-rs`

### Debugging Tips

1. **Request Logging**: Log all incoming requests
2. **Response Validation**: Validate responses before sending
3. **Error Tracking**: Track and categorize errors
4. **SSE Debugging**: Use browser DevTools for SSE streams
5. **Protocol Version**: Always check protocol compatibility