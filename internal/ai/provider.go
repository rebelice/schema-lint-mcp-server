package ai

import (
	"context"
	"fmt"
	"log"
)

// Provider represents an AI provider interface
type Provider interface {
	// Analyze sends a prompt to the AI and returns the response
	Analyze(ctx context.Context, prompt string) (string, error)
	// Name returns the provider name
	Name() string
}

// ProviderFactory creates AI providers based on configuration
type ProviderFactory struct {
	providers map[string]Provider
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers a provider
func (f *ProviderFactory) RegisterProvider(name string, provider Provider) {
	log.Printf("[ProviderFactory] Registering provider: %s", name)
	f.providers[name] = provider
}

// GetProvider returns a provider by name
func (f *ProviderFactory) GetProvider(name string) (Provider, error) {
	log.Printf("[ProviderFactory] Getting provider: %s", name)
	provider, exists := f.providers[name]
	if !exists {
		log.Printf("[ProviderFactory] ERROR: Provider %s not found. Available providers: %v", name, f.ListProviders())
		return nil, fmt.Errorf("provider %s not found", name)
	}
	log.Printf("[ProviderFactory] Provider %s found", name)
	return provider, nil
}

// ListProviders returns all registered provider names
func (f *ProviderFactory) ListProviders() []string {
	var names []string
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}