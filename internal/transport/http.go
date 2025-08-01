package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	mcpserver "github.com/rebeliceyang/schema-lint-mcp-server/internal/server"
)

// HTTPTransport implements MCP HTTP transport with SSE support
type HTTPTransport struct {
	server        *mcpserver.Server
	sseClients    map[string]*sseClient
	mu            sync.RWMutex
	sessionStore  *SessionStore
}

// sseClient represents a Server-Sent Events client connection
type sseClient struct {
	id       string
	writer   http.ResponseWriter
	flusher  http.Flusher
	close    chan bool
	messages chan []byte
}

// SessionStore manages MCP sessions
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// Session represents an MCP session
type Session struct {
	ID              string
	ProtocolVersion string
	ClientInfo      interface{}
	InitializedAt   time.Time
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(srv *mcpserver.Server) *HTTPTransport {
	return &HTTPTransport{
		server:       srv,
		sseClients:   make(map[string]*sseClient),
		sessionStore: NewSessionStore(),
	}
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

// HandleRoot handles GET / requests
func (h *HTTPTransport) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"name":        "sql-schema-lint",
		"version":     "1.0.0",
		"description": "AI-powered MCP server for SQL schema linting",
		"transport":   "http+sse",
		"endpoints": map[string]string{
			"sse":     "/sse",
			"message": "/message",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleHealthz handles GET /healthz requests
func (h *HTTPTransport) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "sql-schema-lint",
		"version": "1.0.0",
		"uptime":  "running",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleSSE handles GET /sse requests for Server-Sent Events
func (h *HTTPTransport) HandleSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if the ResponseWriter supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable Nginx buffering

	// Create SSE client
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	client := &sseClient{
		id:       clientID,
		writer:   w,
		flusher:  flusher,
		close:    make(chan bool),
		messages: make(chan []byte, 100),
	}

	// Register client
	h.mu.Lock()
	h.sseClients[clientID] = client
	h.mu.Unlock()

	// Send initial connection event
	fmt.Fprintf(w, "event: connection\ndata: {\"clientId\":\"%s\"}\n\n", clientID)
	flusher.Flush()

	// Handle client disconnection
	defer func() {
		h.mu.Lock()
		delete(h.sseClients, clientID)
		h.mu.Unlock()
		close(client.close)
		close(client.messages)
	}()

	// Send periodic ping to keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Main event loop
	for {
		select {
		case <-r.Context().Done():
			return
		case <-client.close:
			return
		case msg := <-client.messages:
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, "event: ping\ndata: {\"timestamp\":%d}\n\n", time.Now().Unix())
			flusher.Flush()
		}
	}
}

// HandleMessage handles POST /message requests
func (h *HTTPTransport) HandleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check Accept header for SSE support
	acceptSSE := false
	acceptHeader := r.Header.Get("Accept")
	if acceptHeader != "" && contains(acceptHeader, "text/event-stream") {
		acceptSSE = true
	}

	// Parse JSON-RPC request
	var req json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONRPCError(w, nil, -32700, "Parse error", err.Error())
		return
	}

	// Check if it's a batch request
	var isBatch bool
	if len(req) > 0 && req[0] == '[' {
		isBatch = true
	}

	if isBatch {
		// Handle batch request
		var requests []jsonRPCRequest
		if err := json.Unmarshal(req, &requests); err != nil {
			sendJSONRPCError(w, nil, -32700, "Parse error", err.Error())
			return
		}

		responses := make([]interface{}, 0, len(requests))
		for _, request := range requests {
			if resp := h.handleSingleRequest(r.Context(), request, acceptSSE); resp != nil {
				responses = append(responses, resp)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses)
	} else {
		// Handle single request
		var request jsonRPCRequest
		if err := json.Unmarshal(req, &request); err != nil {
			sendJSONRPCError(w, nil, -32700, "Parse error", err.Error())
			return
		}

		response := h.handleSingleRequest(r.Context(), request, acceptSSE)
		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			// No response for notifications
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// handleSingleRequest processes a single JSON-RPC request
func (h *HTTPTransport) handleSingleRequest(ctx context.Context, req jsonRPCRequest, acceptSSE bool) interface{} {
	// Route to appropriate handler based on method
	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, req)
	case "tools/list":
		return h.handleToolsList(ctx, req)
	case "tools/call":
		return h.handleToolsCall(ctx, req)
	default:
		if req.ID != nil {
			return jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &jsonRPCError{
					Code:    -32601,
					Message: "Method not found",
				},
			}
		}
		return nil
	}
}

// handleInitialize handles the initialize method
func (h *HTTPTransport) handleInitialize(ctx context.Context, req jsonRPCRequest) interface{} {
	var params map[string]interface{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &jsonRPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	// Extract protocol version and client info
	protocolVersion, _ := params["protocolVersion"].(string)
	clientInfo := params["clientInfo"]

	// Create session
	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
	session := &Session{
		ID:              sessionID,
		ProtocolVersion: protocolVersion,
		ClientInfo:      clientInfo,
		InitializedAt:   time.Now(),
	}
	h.sessionStore.Set(sessionID, session)

	// Initialize MCP server session
	result := map[string]interface{}{
		"protocolVersion": protocolVersion,
		"serverInfo": map[string]interface{}{
			"name":    "sql-schema-lint",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
	}

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleToolsList handles the tools/list method
func (h *HTTPTransport) handleToolsList(ctx context.Context, req jsonRPCRequest) interface{} {
	tools := h.server.GetTools()
	
	result := map[string]interface{}{
		"tools": tools,
	}

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleToolsCall handles the tools/call method
func (h *HTTPTransport) handleToolsCall(ctx context.Context, req jsonRPCRequest) interface{} {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &jsonRPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	// Execute tool with the params as a map
	paramsMap := map[string]interface{}{
		"name":      params.Name,
		"arguments": params.Arguments,
	}
	result, err := h.server.CallTool(ctx, paramsMap)
	if err != nil {
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &jsonRPCError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// BroadcastToSSE sends a message to all SSE clients
func (h *HTTPTransport) BroadcastToSSE(event string, data interface{}) {
	msg, err := json.Marshal(data)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients := make([]*sseClient, 0, len(h.sseClients))
	for _, client := range h.sseClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.messages <- msg:
		default:
			// Client buffer full, skip
		}
	}
}

// Session store methods
func (s *SessionStore) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[id]
	return session, exists
}

func (s *SessionStore) Set(id string, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = session
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

// Helper types and functions

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string         `json:"jsonrpc"`
	Result  interface{}    `json:"result,omitempty"`
	Error   *jsonRPCError  `json:"error,omitempty"`
	ID      interface{}    `json:"id"`
}

type jsonRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func sendJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	response := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &jsonRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}