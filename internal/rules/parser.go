package rules

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/rebeliceyang/schema-lint-mcp-server/pkg/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type RuleParser struct {
	markdown goldmark.Markdown
}

func NewRuleParser() *RuleParser {
	return &RuleParser{
		markdown: goldmark.New(),
	}
}

func (p *RuleParser) ParseFile(filePath string) ([]*models.Rule, error) {
	log.Printf("[RuleParser] Parsing rules file: %s", filePath)
	
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[RuleParser] ERROR: Failed to open rules file: %v", err)
		return nil, fmt.Errorf("failed to open rules file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[RuleParser] ERROR: Failed to read rules file: %v", err)
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	log.Printf("[RuleParser] File read successfully, size: %d bytes", len(content))
	return p.Parse(string(content))
}

func (p *RuleParser) Parse(content string) ([]*models.Rule, error) {
	log.Printf("[RuleParser] Starting to parse rules content (length: %d)", len(content))
	
	reader := text.NewReader([]byte(content))
	doc := p.markdown.Parser().Parse(reader)

	var rules []*models.Rule
	var currentRule *models.Rule
	var inRuleSection bool
	var inGoodExample, inBadExample bool
	var exampleContent strings.Builder

	// Walk through the AST
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			// Handle code block exits
			if _, ok := n.(*ast.CodeBlock); ok {
				if inGoodExample && currentRule != nil {
					currentRule.GoodExample = strings.TrimSpace(exampleContent.String())
					inGoodExample = false
					exampleContent.Reset()
				} else if inBadExample && currentRule != nil {
					currentRule.BadExample = strings.TrimSpace(exampleContent.String())
					inBadExample = false
					exampleContent.Reset()
				}
			}
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			if node.Level == 2 {
				// Check if this is a rule heading
				headingText := p.extractText(node, content)
				if strings.HasPrefix(headingText, "Rule:") {
					// Save previous rule if exists
					if currentRule != nil && currentRule.ID != "" {
						rules = append(rules, currentRule)
					}

					// Extract rule ID
					ruleID := strings.TrimSpace(strings.TrimPrefix(headingText, "Rule:"))
					currentRule = &models.Rule{
						ID:   ruleID,
						Tags: []string{},
					}
					inRuleSection = true
					inGoodExample = false
					inBadExample = false
				}
			} else if node.Level == 3 && currentRule != nil {
				// Check for example sections
				headingText := p.extractText(node, content)
				if strings.Contains(headingText, "Good Example") {
					inGoodExample = true
					inBadExample = false
				} else if strings.Contains(headingText, "Bad Example") {
					inBadExample = true
					inGoodExample = false
				}
			}

		case *ast.List:
			if inRuleSection && currentRule != nil {
				// Parse rule metadata from list items
				p.parseRuleMetadata(node, currentRule, content)
			}

		case *ast.Paragraph:
			if inRuleSection && currentRule != nil && currentRule.Description == "" {
				// First paragraph after rule metadata is the description
				if !p.isInList(n) {
					currentRule.Description = strings.TrimSpace(p.extractText(node, content))
				}
			}

		case *ast.CodeBlock:
			if inGoodExample || inBadExample {
				// Extract code block content
				lines := node.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					exampleContent.Write(line.Value([]byte(content)))
				}
			}
		}

		return ast.WalkContinue, nil
	})

	// Don't forget the last rule
	if currentRule != nil && currentRule.ID != "" {
		rules = append(rules, currentRule)
	}

	return rules, nil
}

func (p *RuleParser) parseRuleMetadata(list *ast.List, rule *models.Rule, content string) {
	ast.Walk(list, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if item, ok := n.(*ast.ListItem); ok {
			itemText := p.extractText(item, content)
			p.parseMetadataItem(itemText, rule)
		}

		return ast.WalkContinue, nil
	})
}

func (p *RuleParser) parseMetadataItem(item string, rule *models.Rule) {
	// Remove markdown formatting
	item = strings.ReplaceAll(item, "**", "")
	item = strings.TrimSpace(item)

	// Parse key-value pairs
	colonIndex := strings.Index(item, ":")
	if colonIndex == -1 {
		return
	}

	key := strings.TrimSpace(item[:colonIndex])
	value := strings.TrimSpace(item[colonIndex+1:])

	switch strings.ToLower(key) {
	case "severity":
		rule.Severity = models.Severity(strings.ToLower(value))
	case "target":
		rule.Target = value
	case "check":
		rule.Check = value
	case "tags":
		// Parse comma-separated tags
		tags := strings.Split(value, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				rule.Tags = append(rule.Tags, tag)
			}
		}
	}
}

func (p *RuleParser) extractText(node ast.Node, source string) string {
	var text strings.Builder

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindText:
			if tn, ok := n.(*ast.Text); ok {
				text.Write(tn.Value([]byte(source)))
			}
		case ast.KindCodeSpan:
			if cs, ok := n.(*ast.CodeSpan); ok {
				text.Write(cs.Text([]byte(source)))
			}
		}

		return ast.WalkContinue, nil
	})

	return strings.TrimSpace(text.String())
}

func (p *RuleParser) isInList(node ast.Node) bool {
	parent := node.Parent()
	for parent != nil {
		if _, ok := parent.(*ast.List); ok {
			return true
		}
		parent = parent.Parent()
	}
	return false
}

// ValidateRule checks if a rule has all required fields
func ValidateRule(rule *models.Rule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if rule.Severity == "" {
		return fmt.Errorf("rule severity is required")
	}
	if rule.Target == "" {
		return fmt.Errorf("rule target is required")
	}
	if rule.Check == "" {
		return fmt.Errorf("rule check is required")
	}
	if rule.Description == "" {
		return fmt.Errorf("rule description is required")
	}

	// Validate severity
	switch rule.Severity {
	case models.SeverityError, models.SeverityWarning, models.SeverityInfo:
		// Valid
	default:
		return fmt.Errorf("invalid severity: %s", rule.Severity)
	}

	return nil
}

// RuleSet represents a collection of rules
type RuleSet struct {
	Rules map[string]*models.Rule
}

// NewRuleSet creates a new rule set from parsed rules
func NewRuleSet(rules []*models.Rule) (*RuleSet, error) {
	rs := &RuleSet{
		Rules: make(map[string]*models.Rule),
	}

	for _, rule := range rules {
		if err := ValidateRule(rule); err != nil {
			return nil, fmt.Errorf("invalid rule %s: %w", rule.ID, err)
		}
		if _, exists := rs.Rules[rule.ID]; exists {
			return nil, fmt.Errorf("duplicate rule ID: %s", rule.ID)
		}
		rs.Rules[rule.ID] = rule
	}

	return rs, nil
}

// GetRulesByTarget returns all rules that match a specific target pattern
func (rs *RuleSet) GetRulesByTarget(target string) []*models.Rule {
	var matchingRules []*models.Rule

	for _, rule := range rs.Rules {
		if matchesTarget(rule.Target, target) {
			matchingRules = append(matchingRules, rule)
		}
	}

	return matchingRules
}

// matchesTarget checks if a rule target pattern matches a specific target
func matchesTarget(pattern, target string) bool {
	// Simple pattern matching for now
	// Supports:
	// - Exact match: "tables.users"
	// - Wildcard: "tables.*", "tables.*.columns.*"
	// - Property match: "tables.*.columns[type=varchar]"

	// Convert pattern to regex
	regexPattern := strings.ReplaceAll(pattern, ".", "\\.")
	regexPattern = strings.ReplaceAll(regexPattern, "*", "[^.]+")
	
	// Handle property selectors
	if strings.Contains(regexPattern, "[") && strings.Contains(regexPattern, "]") {
		// For now, ignore property selectors in matching
		regexPattern = regexp.MustCompile(`\[[^\]]+\]`).ReplaceAllString(regexPattern, "")
	}

	regexPattern = "^" + regexPattern + "$"
	
	matched, err := regexp.MatchString(regexPattern, target)
	if err != nil {
		return false
	}

	return matched
}