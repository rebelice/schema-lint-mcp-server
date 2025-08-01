package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/rebeliceyang/schema-lint-mcp-server/internal/config"
	mcpserver "github.com/rebeliceyang/schema-lint-mcp-server/internal/server"
	"github.com/rebeliceyang/schema-lint-mcp-server/internal/transport"
)

func main() {
	// Enable debug logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting SQL Schema Lint MCP Server...")

	// Load configuration
	log.Println("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded successfully: %+v", cfg.Server)

	// Get transport type from environment
	transportType := os.Getenv("MCP_TRANSPORT")
	if transportType == "" {
		transportType = "stdio" // Default to stdio transport
	}
	log.Printf("Using transport type: %s", transportType)

	// Create MCP server
	log.Println("Creating MCP server instance...")
	srv, err := mcpserver.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}
	log.Println("MCP server created successfully")

	// Start the appropriate transport
	switch transportType {
	case "stdio":
		log.Println("Starting MCP server with stdio transport...")
		if err := server.ServeStdio(srv.GetMCPServer()); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case "http":
		// Get port from environment or use default
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		log.Printf("Starting MCP server with HTTP transport on port %s...", port)
		
		// Create HTTP transport
		httpTransport := transport.NewHTTPTransport(srv)
		
		// Setup HTTP routes
		mux := http.NewServeMux()
		mux.HandleFunc("/", httpTransport.HandleRoot)
		mux.HandleFunc("/healthz", httpTransport.HandleHealthz)
		mux.HandleFunc("/sse", httpTransport.HandleSSE)
		mux.HandleFunc("/message", httpTransport.HandleMessage)
		
		// Add CORS middleware
		handler := corsMiddleware(mux)
		
		// Start HTTP server
		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: handler,
		}
		
		log.Printf("HTTP server listening on http://localhost:%s", port)
		log.Printf("SSE endpoint: http://localhost:%s/sse", port)
		log.Printf("Message endpoint: http://localhost:%s/message", port)
		
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	default:
		log.Fatalf("Unknown transport: %s. Use 'stdio' or 'http'", transportType)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Last-Event-ID, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}