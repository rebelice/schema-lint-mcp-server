package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Name        string `mapstructure:"name"`
		Version     string `mapstructure:"version"`
		Description string `mapstructure:"description"`
	} `mapstructure:"server"`

	AI struct {
		DefaultProvider string                    `mapstructure:"default_provider"`
		Providers       map[string]ProviderConfig `mapstructure:"providers"`
		Cache           CacheConfig               `mapstructure:"cache"`
		RateLimiting    RateLimitingConfig        `mapstructure:"rate_limiting"`
	} `mapstructure:"ai"`

	Defaults struct {
		OutputFormat       string `mapstructure:"output_format"`
		SeverityThreshold  string `mapstructure:"severity_threshold"`
		Dialect            string `mapstructure:"dialect"`
		Provider           string `mapstructure:"provider"`
	} `mapstructure:"defaults"`

	Parsers struct {
		SQL      SQLParserConfig                `mapstructure:"sql"`
		Dialects map[string]DialectParserConfig `mapstructure:",remain"`
	} `mapstructure:"parsers"`

	Rules struct {
		BatchSize       int `mapstructure:"batch_size"`
		TimeoutSeconds  int `mapstructure:"timeout_seconds"`
	} `mapstructure:"rules"`

	// Convenience fields
	DefaultProvider     string
	DefaultDialect      string
	DefaultOutputFormat string
}

type ProviderConfig struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type CacheConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	TTLSeconds int  `mapstructure:"ttl_seconds"`
}

type RateLimitingConfig struct {
	MaxRequestsPerMinute int `mapstructure:"max_requests_per_minute"`
	RetryAttempts        int `mapstructure:"retry_attempts"`
	RetryDelayMs         int `mapstructure:"retry_delay_ms"`
}

type SQLParserConfig struct {
	CaseSensitive         bool `mapstructure:"case_sensitive"`
	AllowQuotedIdentifiers bool `mapstructure:"allow_quoted_identifiers"`
	MaxIdentifierLength    int  `mapstructure:"max_identifier_length"`
}

type DialectParserConfig struct {
	Extensions  []string `mapstructure:"extensions"`
	SQLMode     string   `mapstructure:"sql_mode"`
	ForeignKeys bool     `mapstructure:"foreign_keys"`
}

func Load() (*Config, error) {
	v := viper.New()

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/sql-schema-lint/")

	// Set defaults
	setDefaults(v)

	// Bind environment variables
	v.SetEnvPrefix("SQL_LINT")
	v.AutomaticEnv()

	// Override with specific env vars
	bindEnvVars(v)

	// Read config file if exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; use defaults and env vars
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Expand environment variables in API keys
	expandEnvVars(&config)

	// Set convenience fields
	config.DefaultProvider = config.AI.DefaultProvider
	if config.DefaultProvider == "" {
		config.DefaultProvider = config.Defaults.Provider
	}
	config.DefaultDialect = config.Defaults.Dialect
	config.DefaultOutputFormat = config.Defaults.OutputFormat

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.name", "sql-schema-lint")
	v.SetDefault("server.version", "1.0.0")
	v.SetDefault("server.description", "AI-powered MCP server for SQL schema linting")

	// AI defaults
	v.SetDefault("ai.default_provider", "gemini")
	v.SetDefault("ai.cache.enabled", true)
	v.SetDefault("ai.cache.ttl_seconds", 3600)
	v.SetDefault("ai.rate_limiting.max_requests_per_minute", 30)
	v.SetDefault("ai.rate_limiting.retry_attempts", 3)
	v.SetDefault("ai.rate_limiting.retry_delay_ms", 1000)

	// Defaults
	v.SetDefault("defaults.output_format", "json")
	v.SetDefault("defaults.severity_threshold", "warning")
	v.SetDefault("defaults.dialect", "postgres")
	v.SetDefault("defaults.provider", "gemini")

	// Parser defaults
	v.SetDefault("parsers.sql.case_sensitive", false)
	v.SetDefault("parsers.sql.allow_quoted_identifiers", true)
	v.SetDefault("parsers.sql.max_identifier_length", 63)

	// Rules defaults
	v.SetDefault("rules.batch_size", 5)
	v.SetDefault("rules.timeout_seconds", 30)
}

func bindEnvVars(v *viper.Viper) {
	// API Keys
	v.BindEnv("ai.providers.gemini.api_key", "GOOGLE_API_KEY")
	v.BindEnv("ai.providers.claude.api_key", "ANTHROPIC_API_KEY")
	v.BindEnv("ai.providers.openai.api_key", "OPENAI_API_KEY")

	// Model overrides
	v.BindEnv("ai.providers.gemini.model", "GEMINI_MODEL")
	v.BindEnv("ai.providers.claude.model", "CLAUDE_MODEL")
	v.BindEnv("ai.providers.openai.model", "OPENAI_MODEL")
}

// expandEnvVars expands environment variables in config values
func expandEnvVars(config *Config) {
	for name, provider := range config.AI.Providers {
		// Expand environment variables in API keys
		if strings.HasPrefix(provider.APIKey, "${") && strings.HasSuffix(provider.APIKey, "}") {
			envVar := strings.TrimSuffix(strings.TrimPrefix(provider.APIKey, "${"), "}")
			if value := os.Getenv(envVar); value != "" {
				provider.APIKey = value
				config.AI.Providers[name] = provider
			}
		}
	}
}

func (c *Config) Validate() error {
	// Check if at least one provider is configured
	hasProvider := false
	for name, provider := range c.AI.Providers {
		if provider.APIKey != "" {
			hasProvider = true
			if provider.Model == "" {
				// Set default models if not specified
				switch name {
				case "gemini":
					c.AI.Providers[name] = ProviderConfig{
						APIKey:    provider.APIKey,
						Model:     "gemini-2.0-flash-exp",
						MaxTokens: 8192,
					}
				case "claude":
					c.AI.Providers[name] = ProviderConfig{
						APIKey:    provider.APIKey,
						Model:     "claude-3-5-sonnet-20241022",
						MaxTokens: 4096,
					}
				case "openai":
					c.AI.Providers[name] = ProviderConfig{
						APIKey:    provider.APIKey,
						Model:     "gpt-4-turbo-preview",
						MaxTokens: 4096,
					}
				}
			}
		}
	}

	if !hasProvider {
		return fmt.Errorf("at least one AI provider must be configured with an API key. Set GOOGLE_API_KEY, ANTHROPIC_API_KEY, or OPENAI_API_KEY environment variable")
	}

	// Validate default provider
	if _, exists := c.AI.Providers[c.DefaultProvider]; !exists || c.AI.Providers[c.DefaultProvider].APIKey == "" {
		// Find first available provider
		for name, provider := range c.AI.Providers {
			if provider.APIKey != "" {
				c.DefaultProvider = name
				break
			}
		}
	}

	return nil
}

func (c *Config) GetProvider(name string) (*ProviderConfig, error) {
	if name == "" {
		name = c.DefaultProvider
	}

	provider, exists := c.AI.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not configured", name)
	}

	if provider.APIKey == "" {
		return nil, fmt.Errorf("API key not set for provider %s", name)
	}

	return &provider, nil
}

func (c *Config) GetCacheTTL() time.Duration {
	return time.Duration(c.AI.Cache.TTLSeconds) * time.Second
}