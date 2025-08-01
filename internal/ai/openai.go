package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client *openai.Client
	config OpenAIConfig
}

// OpenAIConfig holds configuration for OpenAI provider
type OpenAIConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config OpenAIConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(config.APIKey)

	return &OpenAIProvider{
		client: client,
		config: config,
	}, nil
}

// Analyze sends a prompt to OpenAI and returns the response
func (p *OpenAIProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	model := p.config.Model
	if model == "" {
		model = openai.GPT4TurboPreview
	}

	maxTokens := p.config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert SQL schema analyzer. Analyze the given schema against the provided rules and return structured feedback.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.1, // Low temperature for consistent analysis
			MaxTokens:   maxTokens,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI analysis failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}