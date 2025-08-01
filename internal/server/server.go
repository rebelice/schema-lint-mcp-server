package server

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rebeliceyang/schema-lint-mcp-server/internal/config"
	"github.com/rebeliceyang/schema-lint-mcp-server/internal/lint"
)

type Server struct {
	mcpServer *server.MCPServer
	config    *config.Config
	linter    *lint.Linter
}

func NewServer(cfg *config.Config) (*Server, error) {
	log.Println("[Server] Creating new server instance...")
	
	// Create linter
	log.Println("[Server] Creating linter...")
	linter, err := lint.NewLinter(cfg)
	if err != nil {
		log.Printf("[Server] ERROR: Failed to create linter: %v", err)
		return nil, fmt.Errorf("failed to create linter: %w", err)
	}
	log.Println("[Server] Linter created successfully")

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"sql-schema-lint",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s := &Server{
		mcpServer: mcpServer,
		config:    cfg,
		linter:    linter,
	}

	// Register tools
	log.Println("[Server] Registering tools...")
	if err := s.registerTools(); err != nil {
		log.Printf("[Server] ERROR: Failed to register tools: %v", err)
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}
	log.Println("[Server] Tools registered successfully")

	return s, nil
}

func (s *Server) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}

// GetTools returns all registered tools
func (s *Server) GetTools() []interface{} {
	// For now, return the lint_schema tool
	// In a real implementation, this would query the mcpServer for registered tools
	return []interface{}{
		map[string]interface{}{
			"name":        "lint_schema",
			"description": "Validate SQL schema against user-defined rules using AI",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"schema_context": map[string]interface{}{
						"type":        "string",
						"description": "SQL schema content to be validated",
					},
					"rules_context": map[string]interface{}{
						"type":        "string",
						"description": "Rules content in markdown format",
					},
					"provider": map[string]interface{}{
						"type":        "string",
						"description": "AI provider to use (gemini, claude, openai)",
						"enum":        []string{"gemini", "claude", "openai"},
					},
					"dialect": map[string]interface{}{
						"type":        "string",
						"description": "SQL dialect (postgres, mysql, sqlite, sqlserver, oracle)",
						"enum":        []string{"postgres", "mysql", "sqlite", "sqlserver", "oracle"},
					},
					"output_format": map[string]interface{}{
						"type":        "string",
						"description": "Output format for lint results",
						"enum":        []string{"json", "text", "markdown"},
					},
				},
				"required": []string{"schema_path", "rules_path"},
			},
		},
	}
}

// CallTool executes a tool by name
func (s *Server) CallTool(ctx context.Context, req interface{}) (*mcp.CallToolResult, error) {
	log.Printf("[Server] CallTool invoked with request: %+v", req)
	
	// Extract tool name and arguments from the request
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		log.Printf("[Server] ERROR: Invalid request format: %T", req)
		return nil, fmt.Errorf("invalid request format")
	}
	
	name, ok := reqMap["name"].(string)
	if !ok {
		log.Printf("[Server] ERROR: Tool name not found in request: %+v", reqMap)
		return nil, fmt.Errorf("tool name not found")
	}
	
	log.Printf("[Server] Tool name: %s", name)
	
	if name == "lint_schema" {
		// Pass the arguments directly to the handler
		args, ok := reqMap["arguments"].(map[string]interface{})
		if !ok {
			log.Printf("[Server] ERROR: Invalid arguments format: %T", reqMap["arguments"])
			return nil, fmt.Errorf("invalid arguments format")
		}
		log.Printf("[Server] Calling lint_schema with args: %+v", args)
		return s.handleLintSchemaWithArgs(ctx, args)
	}
	log.Printf("[Server] ERROR: Unknown tool: %s", name)
	return nil, fmt.Errorf("unknown tool: %s", name)
}

func (s *Server) registerTools() error {
	// Create lint_schema tool
	lintSchemaTool := mcp.NewTool("lint_schema",
		mcp.WithDescription("Validate SQL schema against user-defined rules using AI"),
		mcp.WithString("schema_context",
			mcp.Required(),
			mcp.Description("SQL schema content to be validated"),
		),
		mcp.WithString("rules_context",
			mcp.Required(),
			mcp.Description("Rules content in markdown format"),
		),
		mcp.WithString("provider",
			mcp.Description("AI provider to use (gemini, claude, openai)"),
			mcp.Enum("gemini", "claude", "openai"),
		),
		mcp.WithString("dialect",
			mcp.Description("SQL dialect (postgres, mysql, sqlite, sqlserver, oracle)"),
			mcp.Enum("postgres", "mysql", "sqlite", "sqlserver", "oracle"),
		),
		mcp.WithString("output_format",
			mcp.Description("Output format for lint results"),
			mcp.Enum("json", "text", "markdown"),
		),
	)

	// Add tool with handler
	s.mcpServer.AddTool(lintSchemaTool, s.handleLintSchema)

	return nil
}

func (s *Server) handleLintSchemaWithArgs(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	log.Println("[Server] handleLintSchemaWithArgs called")
	log.Printf("[Server] Arguments: %+v", args)
	
	// Extract required arguments
	schemaContext, ok := args["schema_context"].(string)
	if !ok {
		log.Printf("[Server] ERROR: schema_context not found or invalid type: %v", args["schema_context"])
		return mcp.NewToolResultError("schema_context is required"), nil
	}
	log.Printf("[Server] Schema context length: %d characters", len(schemaContext))

	rulesContext, ok := args["rules_context"].(string)
	if !ok {
		log.Printf("[Server] ERROR: rules_context not found or invalid type: %v", args["rules_context"])
		return mcp.NewToolResultError("rules_context is required"), nil
	}
	log.Printf("[Server] Rules context length: %d characters", len(rulesContext))

	// Get optional arguments
	provider := s.config.DefaultProvider
	if p, ok := args["provider"].(string); ok && p != "" {
		provider = p
	}

	dialect := s.config.DefaultDialect
	if d, ok := args["dialect"].(string); ok && d != "" {
		dialect = d
	}

	outputFormat := s.config.DefaultOutputFormat
	if f, ok := args["output_format"].(string); ok && f != "" {
		outputFormat = f
	}

	// Create lint request
	lintRequest := &lint.Request{
		SchemaContent: schemaContext,
		RulesContent:  rulesContext,
		Provider:      provider,
		Dialect:       dialect,
		OutputFormat:  outputFormat,
	}

	// Perform linting
	log.Printf("[Server] Starting lint with request: %+v", lintRequest)
	result, err := s.linter.Lint(ctx, lintRequest)
	if err != nil {
		log.Printf("[Server] ERROR: Linting failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Linting failed: %v", err)), nil
	}
	log.Printf("[Server] Linting completed successfully")

	// Format output based on requested format
	var output string
	switch outputFormat {
	case "json":
		output, err = result.ToJSON()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format JSON output: %v", err)), nil
		}
	case "text":
		output = result.ToText()
	case "markdown":
		output = result.ToMarkdown()
	default:
		output, err = result.ToJSON()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format output: %v", err)), nil
		}
	}

	return mcp.NewToolResultText(output), nil
}

func (s *Server) handleLintSchema(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required arguments
	schemaContext, err := request.RequireString("schema_context")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("schema_context is required: %v", err)), nil
	}

	rulesContext, err := request.RequireString("rules_context")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("rules_context is required: %v", err)), nil
	}

	// Get optional arguments
	args := request.GetArguments()
	
	provider := s.config.DefaultProvider
	if p, ok := args["provider"].(string); ok && p != "" {
		provider = p
	}

	dialect := s.config.DefaultDialect
	if d, ok := args["dialect"].(string); ok && d != "" {
		dialect = d
	}

	outputFormat := s.config.DefaultOutputFormat
	if f, ok := args["output_format"].(string); ok && f != "" {
		outputFormat = f
	}

	// Create lint request
	lintRequest := &lint.Request{
		SchemaContent: schemaContext,
		RulesContent:  rulesContext,
		Provider:      provider,
		Dialect:       dialect,
		OutputFormat:  outputFormat,
	}

	// Perform linting
	log.Printf("[Server] Starting lint with request: %+v", lintRequest)
	result, err := s.linter.Lint(ctx, lintRequest)
	if err != nil {
		log.Printf("[Server] ERROR: Linting failed: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("Linting failed: %v", err)), nil
	}
	log.Printf("[Server] Linting completed successfully")

	// Format output based on requested format
	var output string
	switch outputFormat {
	case "json":
		output, err = result.ToJSON()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format JSON output: %v", err)), nil
		}
	case "text":
		output = result.ToText()
	case "markdown":
		output = result.ToMarkdown()
	default:
		output, err = result.ToJSON()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format output: %v", err)), nil
		}
	}

	return mcp.NewToolResultText(output), nil
}