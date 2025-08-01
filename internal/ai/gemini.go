package ai

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	client *genai.Client
	model  *genai.GenerativeModel
	config GeminiConfig
}

// GeminiConfig holds configuration for Gemini provider
type GeminiConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(ctx context.Context, config GeminiConfig) (*GeminiProvider, error) {
	log.Printf("[Gemini] Creating new Gemini provider with model: %s", config.Model)
	
	if config.APIKey == "" {
		log.Println("[Gemini] ERROR: API key is empty")
		return nil, fmt.Errorf("Gemini API key is required")
	}

	log.Println("[Gemini] Creating Gemini client...")
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		log.Printf("[Gemini] ERROR: Failed to create client: %v", err)
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel(config.Model)
	
	// Configure model settings
	model.SetTemperature(0.1) // Low temperature for consistent analysis
	if config.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(config.MaxTokens))
	}

	// Configure safety settings for code analysis
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
	}

	return &GeminiProvider{
		client: client,
		model:  model,
		config: config,
	}, nil
}

// Analyze sends a prompt to Gemini and returns the response
func (p *GeminiProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	log.Printf("[Gemini] Analyzing prompt (length: %d chars)", len(prompt))
	
	resp, err := p.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("[Gemini] ERROR: Analysis failed: %v", err)
		return "", fmt.Errorf("Gemini analysis failed: %w", err)
	}

	if len(resp.Candidates) == 0 {
		log.Println("[Gemini] ERROR: No response candidates from Gemini")
		return "", fmt.Errorf("no response from Gemini")
	}

	// Extract text from response
	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}

	log.Printf("[Gemini] Analysis completed successfully (response length: %d chars)", len(result))
	return result, nil
}

// Name returns the provider name
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// Close closes the Gemini client
func (p *GeminiProvider) Close() error {
	return p.client.Close()
}