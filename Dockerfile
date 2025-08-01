# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o mcp-server cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mcp-server .

# Copy example files
COPY --from=builder /app/examples ./examples

# Copy config file if exists
COPY --from=builder /app/config.yaml* ./

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (for HTTP transport)
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV MCP_TRANSPORT=stdio

# Run the server
CMD ["./mcp-server"]