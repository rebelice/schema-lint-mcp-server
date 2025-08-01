package lint

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/rebeliceyang/schema-lint-mcp-server/internal/ai"
	"github.com/rebeliceyang/schema-lint-mcp-server/internal/config"
	"github.com/rebeliceyang/schema-lint-mcp-server/internal/parser"
	"github.com/rebeliceyang/schema-lint-mcp-server/internal/rules"
	"github.com/rebeliceyang/schema-lint-mcp-server/pkg/models"
	"golang.org/x/time/rate"
)

// Request represents a lint request
type Request struct {
	SchemaPath    string // Deprecated: Use SchemaContent instead
	RulesPath     string // Deprecated: Use RulesContent instead
	SchemaContent string // SQL schema content
	RulesContent  string // Rules content in markdown format
	Provider      string
	Dialect       string
	OutputFormat  string
}

// Linter performs SQL schema linting using AI
type Linter struct {
	config          *config.Config
	sqlParser       *parser.SQLParser
	ruleParser      *rules.RuleParser
	promptBuilder   *ai.PromptBuilder
	providerFactory *ai.ProviderFactory
	cache           *Cache
	rateLimiters    map[string]*rate.Limiter
	mu              sync.Mutex
}

// NewLinter creates a new linter instance
func NewLinter(cfg *config.Config) (*Linter, error) {
	log.Println("[Linter] Creating new linter instance...")
	l := &Linter{
		config:          cfg,
		ruleParser:      rules.NewRuleParser(),
		promptBuilder:   ai.NewPromptBuilder(),
		providerFactory: ai.NewProviderFactory(),
		rateLimiters:    make(map[string]*rate.Limiter),
	}

	// Initialize cache if enabled
	if cfg.AI.Cache.Enabled {
		log.Println("[Linter] Cache enabled, initializing...")
		l.cache = NewCache(cfg.GetCacheTTL())
	}

	// Initialize AI providers
	log.Println("[Linter] Initializing AI providers...")
	if err := l.initializeProviders(); err != nil {
		log.Printf("[Linter] ERROR: Failed to initialize AI providers: %v", err)
		return nil, fmt.Errorf("failed to initialize AI providers: %w", err)
	}

	return l, nil
}

func (l *Linter) initializeProviders() error {
	ctx := context.Background()
	providersInitialized := 0

	// Initialize Gemini provider
	if geminiCfg, err := l.config.GetProvider("gemini"); err == nil && geminiCfg.APIKey != "" {
		log.Printf("[Linter] Initializing Gemini provider with model: %s", geminiCfg.Model)
		provider, err := ai.NewGeminiProvider(ctx, ai.GeminiConfig{
			APIKey:    geminiCfg.APIKey,
			Model:     geminiCfg.Model,
			MaxTokens: geminiCfg.MaxTokens,
		})
		if err == nil {
			l.providerFactory.RegisterProvider("gemini", provider)
			l.rateLimiters["gemini"] = rate.NewLimiter(
				rate.Every(time.Minute/time.Duration(l.config.AI.RateLimiting.MaxRequestsPerMinute)),
				1,
			)
			providersInitialized++
			log.Println("[Linter] Gemini provider initialized successfully")
		} else {
			log.Printf("[Linter] WARNING: Failed to initialize Gemini provider: %v", err)
		}
	} else {
		log.Println("[Linter] Gemini provider not configured or API key missing")
	}

	// Initialize Claude provider
	if claudeCfg, err := l.config.GetProvider("claude"); err == nil && claudeCfg.APIKey != "" {
		provider, err := ai.NewClaudeProvider(ai.ClaudeConfig{
			APIKey:    claudeCfg.APIKey,
			Model:     claudeCfg.Model,
			MaxTokens: claudeCfg.MaxTokens,
		})
		if err == nil {
			l.providerFactory.RegisterProvider("claude", provider)
			l.rateLimiters["claude"] = rate.NewLimiter(
				rate.Every(time.Minute/time.Duration(l.config.AI.RateLimiting.MaxRequestsPerMinute)),
				1,
			)
			providersInitialized++
		} else {
			fmt.Printf("Warning: Failed to initialize Claude provider: %v\n", err)
		}
	}

	// Initialize OpenAI provider
	if openaiCfg, err := l.config.GetProvider("openai"); err == nil && openaiCfg.APIKey != "" {
		provider, err := ai.NewOpenAIProvider(ai.OpenAIConfig{
			APIKey:    openaiCfg.APIKey,
			Model:     openaiCfg.Model,
			MaxTokens: openaiCfg.MaxTokens,
		})
		if err == nil {
			l.providerFactory.RegisterProvider("openai", provider)
			l.rateLimiters["openai"] = rate.NewLimiter(
				rate.Every(time.Minute/time.Duration(l.config.AI.RateLimiting.MaxRequestsPerMinute)),
				1,
			)
			providersInitialized++
		} else {
			fmt.Printf("Warning: Failed to initialize OpenAI provider: %v\n", err)
		}
	}

	// Check if at least one provider is available
	if providersInitialized == 0 {
		return fmt.Errorf("no AI providers were successfully initialized. Please set one of: GOOGLE_API_KEY, ANTHROPIC_API_KEY, or OPENAI_API_KEY")
	}

	return nil
}

// Lint performs linting on a SQL schema file
func (l *Linter) Lint(ctx context.Context, req *Request) (*models.LintResult, error) {
	log.Printf("[Linter] Starting lint operation with request: %+v", req)
	
	// Initialize SQL parser with dialect
	log.Printf("[Linter] Initializing SQL parser with dialect: %s", req.Dialect)
	l.sqlParser = parser.NewSQLParser(req.Dialect)

	// Parse SQL schema
	var schema *models.Schema
	var err error
	if req.SchemaContent != "" {
		log.Printf("[Linter] Parsing SQL schema from content (length: %d)", len(req.SchemaContent))
		schema, err = l.sqlParser.Parse(req.SchemaContent)
	} else if req.SchemaPath != "" {
		log.Printf("[Linter] Parsing SQL schema file: %s", req.SchemaPath)
		schema, err = l.sqlParser.ParseFile(req.SchemaPath)
	} else {
		return nil, fmt.Errorf("either SchemaContent or SchemaPath must be provided")
	}
	if err != nil {
		log.Printf("[Linter] ERROR: Failed to parse SQL schema: %v", err)
		return nil, fmt.Errorf("failed to parse SQL schema: %w", err)
	}
	log.Printf("[Linter] SQL schema parsed successfully, found %d objects", len(schema.Tables))

	// Parse rules
	var ruleList []*models.Rule
	if req.RulesContent != "" {
		log.Printf("[Linter] Parsing rules from content (length: %d)", len(req.RulesContent))
		ruleList, err = l.ruleParser.Parse(req.RulesContent)
	} else if req.RulesPath != "" {
		log.Printf("[Linter] Parsing rules file: %s", req.RulesPath)
		ruleList, err = l.ruleParser.ParseFile(req.RulesPath)
	} else {
		return nil, fmt.Errorf("either RulesContent or RulesPath must be provided")
	}
	if err != nil {
		log.Printf("[Linter] ERROR: Failed to parse rules: %v", err)
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}
	log.Printf("[Linter] Rules parsed successfully, found %d rules", len(ruleList))

	// Create rule set for validation
	_, err = rules.NewRuleSet(ruleList)
	if err != nil {
		return nil, fmt.Errorf("failed to create rule set: %w", err)
	}

	// Get AI provider
	log.Printf("[Linter] Getting AI provider: %s", req.Provider)
	provider, err := l.providerFactory.GetProvider(req.Provider)
	if err != nil {
		log.Printf("[Linter] ERROR: Failed to get AI provider %s: %v", req.Provider, err)
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}
	log.Printf("[Linter] AI provider %s retrieved successfully", req.Provider)

	// Apply rate limiting
	limiter := l.rateLimiters[req.Provider]
	if limiter != nil {
		if err := limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit exceeded: %w", err)
		}
	}

	// Analyze schema against rules
	var allIssues []models.Issue

	// Batch rules for efficiency
	ruleBatches := l.batchRules(ruleList, l.config.Rules.BatchSize)

	for _, batch := range ruleBatches {
		if len(batch) == 1 {
			// Single rule analysis
			issues, err := l.analyzeRule(ctx, provider, schema, batch[0])
			if err != nil {
				// Log error but continue with other rules
				log.Printf("[Linter] ERROR: Failed to analyze rule %s: %v", batch[0].ID, err)
				continue
			}
			allIssues = append(allIssues, issues...)
		} else {
			// Batch analysis
			batchIssues, err := l.analyzeBatch(ctx, provider, schema, batch)
			if err != nil {
				// Fall back to individual analysis
				for _, rule := range batch {
					issues, err := l.analyzeRule(ctx, provider, schema, rule)
					if err != nil {
						fmt.Printf("Error analyzing rule %s: %v\n", rule.ID, err)
						continue
					}
					allIssues = append(allIssues, issues...)
				}
			} else {
				for _, issues := range batchIssues {
					allIssues = append(allIssues, issues...)
				}
			}
		}
	}

	// Sort issues by line number
	sort.Slice(allIssues, func(i, j int) bool {
		if allIssues[i].Line == allIssues[j].Line {
			return allIssues[i].Column < allIssues[j].Column
		}
		return allIssues[i].Line < allIssues[j].Line
	})

	// Build result
	schemaFile := req.SchemaPath
	if schemaFile == "" && req.SchemaContent != "" {
		schemaFile = "<inline>"
	}
	result := &models.LintResult{
		SchemaFile:  schemaFile,
		Dialect:     schema.Dialect,
		TotalIssues: len(allIssues),
		Issues:      allIssues,
	}

	// Count issues by severity
	for _, issue := range allIssues {
		switch issue.Severity {
		case models.SeverityError:
			result.Errors++
		case models.SeverityWarning:
			result.Warnings++
		case models.SeverityInfo:
			result.Info++
		}
	}

	return result, nil
}

func (l *Linter) analyzeRule(ctx context.Context, provider ai.Provider, schema *models.Schema, rule *models.Rule) ([]models.Issue, error) {
	log.Printf("[Linter] Analyzing rule: %s (%s)", rule.ID, rule.Description)
	
	// Check cache first
	if l.cache != nil {
		cacheKey := l.cache.Key(schema.Raw, rule.ID, provider.Name())
		if cached, found := l.cache.Get(cacheKey); found {
			log.Printf("[Linter] Cache hit for rule %s", rule.ID)
			return cached, nil
		}
	}

	// Build prompt
	prompt := l.promptBuilder.BuildRulePrompt(schema, rule)

	// Get AI analysis with retry
	var response string
	var err error
	for i := 0; i < l.config.AI.RateLimiting.RetryAttempts; i++ {
		response, err = provider.Analyze(ctx, prompt)
		if err == nil {
			break
		}
		if i < l.config.AI.RateLimiting.RetryAttempts-1 {
			time.Sleep(time.Duration(l.config.AI.RateLimiting.RetryDelayMs) * time.Millisecond)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("AI analysis failed after %d attempts: %w", l.config.AI.RateLimiting.RetryAttempts, err)
	}

	// Parse response
	issues, err := l.promptBuilder.ParseRuleResponse(response, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Cache result
	if l.cache != nil {
		cacheKey := l.cache.Key(schema.Raw, rule.ID, provider.Name())
		l.cache.Set(cacheKey, issues)
	}

	return issues, nil
}

func (l *Linter) analyzeBatch(ctx context.Context, provider ai.Provider, schema *models.Schema, batch []*models.Rule) (map[string][]models.Issue, error) {
	// Build batch prompt
	prompt := l.promptBuilder.BuildBatchPrompt(schema, batch)

	// Get AI analysis
	response, err := provider.Analyze(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("batch AI analysis failed: %w", err)
	}

	// Parse batch response
	return l.promptBuilder.ParseBatchResponse(response, batch)
}

func (l *Linter) batchRules(rules []*models.Rule, batchSize int) [][]*models.Rule {
	if batchSize <= 0 {
		batchSize = 1
	}

	var batches [][]*models.Rule
	for i := 0; i < len(rules); i += batchSize {
		end := i + batchSize
		if end > len(rules) {
			end = len(rules)
		}
		batches = append(batches, rules[i:end])
	}

	return batches
}