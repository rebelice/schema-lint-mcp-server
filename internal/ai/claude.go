package ai

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeProvider implements the Provider interface for Anthropic Claude
type ClaudeProvider struct {
	client *anthropic.Client
	config ClaudeConfig
}

// ClaudeConfig holds configuration for Claude provider
type ClaudeConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(config ClaudeConfig) (*ClaudeProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(config.APIKey),
	)

	return &ClaudeProvider{
		client: &client,
		config: config,
	}, nil
}

// Analyze sends a prompt to Claude and returns the response
func (p *ClaudeProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	maxTokens := p.config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// Use the model from config or default
	model := anthropic.ModelClaude3_5Sonnet20241022
	if p.config.Model != "" {
		model = anthropic.Model(p.config.Model)
	}

	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: int64(maxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Temperature: anthropic.Float(0.1), // Low temperature for consistent analysis
	})

	if err != nil {
		return "", fmt.Errorf("Claude analysis failed: %w", err)
	}

	// Extract text from response
	var result string
	for _, block := range message.Content {
		// Access the text content directly
		result += string(block.Text)
	}

	return result, nil
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string {
	return "claude"
}