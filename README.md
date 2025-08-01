# SQL Schema Lint MCP Server

An AI-powered MCP (Model Context Protocol) server that provides intelligent SQL schema validation using AI models (Gemini, Claude, OpenAI) to lint SQL schema files against user-defined rules.

## Features

- 🤖 **AI-Powered Analysis**: Leverages state-of-the-art AI models for intelligent schema analysis
- 📝 **Markdown Rule Definition**: Define rules in readable Markdown format
- 🔍 **Multi-Dialect Support**: PostgreSQL, MySQL, SQLite, SQL Server, Oracle
- 🚀 **MCP Protocol**: Full implementation of Model Context Protocol with both stdio and HTTP+SSE transports
- 💾 **Smart Caching**: Cache AI responses for improved performance
- 🎯 **Batch Processing**: Analyze multiple rules in a single AI call
- 📊 **Multiple Output Formats**: JSON, Text, and Markdown output
- 🌐 **HTTP Transport**: RESTful API with Server-Sent Events for real-time communication

## Quick Start

### Prerequisites

- Go 1.21 or later
- At least one AI provider API key:
  - Google Gemini API key
  - Anthropic Claude API key
  - OpenAI API key

### Installation

```bash
# Clone the repository
git clone https://github.com/rebeliceyang/schema-lint-mcp-server.git
cd schema-lint-mcp-server

# Install dependencies
go mod download

# Set up environment variables
export GOOGLE_API_KEY="your-gemini-api-key"
# or
export ANTHROPIC_API_KEY="your-claude-api-key"
# or
export OPENAI_API_KEY="your-openai-api-key"

# Run the server (stdio transport)
go run cmd/server/main.go

# Or run with HTTP transport
MCP_TRANSPORT=http PORT=8080 go run cmd/server/main.go
```

The server supports both stdio and HTTP transports for MCP communication.

### Basic Usage

#### Stdio Transport (Default)

The server implements the MCP protocol with stdio transport by default:

1. **Initialize a session**:
```bash
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
```

2. **List available tools**:
```bash
echo '{
    "jsonrpc": "2.0",
    "method": "tools/list",
    "id": 2
}' | go run cmd/server/main.go
```

3. **Lint a schema**:
```bash
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
```

#### HTTP Transport

To use HTTP transport with Server-Sent Events:

1. **Start the server**:
```bash
./start_http_server.sh
# or manually:
MCP_TRANSPORT=http PORT=8080 go run cmd/server/main.go
```

2. **Test endpoints**:
```bash
# Get server info
curl http://localhost:8080/

# Check health
curl http://localhost:8080/healthz

# Connect to SSE stream
curl -N -H "Accept: text/event-stream" http://localhost:8080/sse
```

3. **Initialize session**:
```bash
curl -X POST http://localhost:8080/message \
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
```

4. **Call tools**:
```bash
curl -X POST http://localhost:8080/message \
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
    "id": 2
  }'
```

### Using with Claude Desktop

Add the server to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "sql-schema-lint": {
      "command": "/path/to/mcp-server",
      "env": {
        "GOOGLE_API_KEY": "your-api-key"
      }
    }
  }
}
```

## Configuration

Create a `config.yaml` file in the project root:

```yaml
ai:
  default_provider: "gemini"
  providers:
    gemini:
      api_key: "${GOOGLE_API_KEY}"
      model: "gemini-2.0-flash-exp"
      max_tokens: 8192
    
defaults:
  output_format: "json"
  dialect: "postgres"
  
rules:
  batch_size: 5
  timeout_seconds: 30
```

## Writing Rules

Rules are defined in Markdown format. Each rule should follow this structure:

```markdown
## Rule: rule-id
- **Severity**: error|warning|info
- **Target**: tables.*|columns.*|etc
- **Check**: check_expression
- **Tags**: optional, tags

Description of what the rule checks.

### Good Example
```sql
-- Example of correct usage
```

### Bad Example
```sql
-- Example that violates the rule
```
```

See `examples/rules/` for more examples.

## API Reference

### MCP Tools

#### `lint_schema`

Validates SQL schema files against custom lint rules.

**Parameters:**
- `schema_path` (string, required): Path to SQL schema file
- `rules_path` (string, required): Path to rules file (.md)
- `provider` (string, optional): AI provider (gemini, claude, openai)
- `dialect` (string, optional): SQL dialect
- `output_format` (string, optional): Output format (json, text, markdown)

## HTTP Transport Endpoints

When running with HTTP transport, the server exposes the following endpoints:

### `GET /`
Returns server information and available endpoints.

### `GET /healthz`
Health check endpoint returning server status.

### `GET /sse`
Server-Sent Events endpoint for real-time bidirectional communication. Clients can establish a persistent connection to receive server notifications.

### `POST /message`
Main JSON-RPC endpoint for MCP protocol messages. Accepts both single requests and batch requests.

**Headers:**
- `Content-Type: application/json`
- `Accept: application/json, text/event-stream` (optional, for SSE responses)

## Architecture

```
┌─────────────────┐     ┌──────────────────┐
│   MCP Client    │────▶│   MCP Server     │
└─────────────────┘     │ (stdio/http+sse) │
                        └──────────────────┘
                               │
                    ┌──────────┴──────────┐
                    │                     │
              ┌─────▼─────┐         ┌─────▼─────┐
              │  Linter   │         │  Config   │
              └─────┬─────┘         └───────────┘
                    │
       ┌────────────┼────────────┐
       │            │            │
  ┌────▼────┐ ┌────▼────┐ ┌────▼────┐
  │  SQL    │ │  Rules  │ │   AI    │
  │ Parser  │ │ Parser  │ │Provider │
  └─────────┘ └─────────┘ └─────────┘
```

## Development

### Project Structure

```
.
├── cmd/server/          # Server entry point
├── internal/            # Internal packages
│   ├── ai/             # AI provider integrations
│   ├── config/         # Configuration management
│   ├── lint/           # Core linting logic
│   ├── parser/         # SQL parser
│   ├── rules/          # Rules parser
│   └── server/         # MCP server implementation
├── pkg/models/         # Shared data models
├── examples/           # Example schemas and rules
└── config.yaml         # Configuration file
```

### Running Tests

```bash
# Run unit tests
go test ./...

# Test stdio transport
./test_stdio.sh

# Test HTTP transport
# Terminal 1: Start server
./start_http_server.sh

# Terminal 2: Run tests
./test_http.sh

# Test SSE endpoint
./test_sse.sh
```

### Building

```bash
go build -o mcp-server cmd/server/main.go
```

## Docker Support

```bash
# Build image
docker build -t sql-schema-lint-mcp .

# Run container with stdio transport
docker run -it \
  -e GOOGLE_API_KEY=your-api-key \
  sql-schema-lint-mcp

# Run container with HTTP transport
docker run -p 8080:8080 \
  -e GOOGLE_API_KEY=your-api-key \
  -e MCP_TRANSPORT=http \
  -e PORT=8080 \
  sql-schema-lint-mcp
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [MCP Go SDK](https://github.com/mark3labs/mcp-go) for the MCP implementation
- Google Gemini, Anthropic Claude, and OpenAI for AI capabilities
- The SQL parsing libraries that make schema analysis possible