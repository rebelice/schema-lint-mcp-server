package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rebeliceyang/schema-lint-mcp-server/pkg/models"
)

// PromptBuilder builds prompts for AI providers
type PromptBuilder struct{}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// BuildRulePrompt creates a prompt for analyzing a schema against a specific rule
func (pb *PromptBuilder) BuildRulePrompt(schema *models.Schema, rule *models.Rule) string {
	var prompt strings.Builder

	prompt.WriteString("Analyze the following SQL schema for compliance with this rule:\n\n")
	
	// Rule details
	prompt.WriteString(fmt.Sprintf("Rule ID: %s\n", rule.ID))
	prompt.WriteString(fmt.Sprintf("Severity: %s\n", rule.Severity))
	prompt.WriteString(fmt.Sprintf("Target: %s\n", rule.Target))
	prompt.WriteString(fmt.Sprintf("Check: %s\n", rule.Check))
	prompt.WriteString(fmt.Sprintf("Description: %s\n\n", rule.Description))

	// Good/Bad examples if provided
	if rule.GoodExample != "" {
		prompt.WriteString("Good Example:\n```sql\n")
		prompt.WriteString(rule.GoodExample)
		prompt.WriteString("\n```\n\n")
	}

	if rule.BadExample != "" {
		prompt.WriteString("Bad Example:\n```sql\n")
		prompt.WriteString(rule.BadExample)
		prompt.WriteString("\n```\n\n")
	}

	// Schema information
	prompt.WriteString(fmt.Sprintf("SQL Dialect: %s\n\n", schema.Dialect))
	prompt.WriteString("SQL Schema:\n```sql\n")
	prompt.WriteString(schema.Raw)
	prompt.WriteString("\n```\n\n")

	// Instructions
	prompt.WriteString("Check if the schema violates this rule. For each violation found, provide:\n")
	prompt.WriteString("1. The specific object (table/column/constraint) that violates the rule\n")
	prompt.WriteString("2. The object type (table, column, index, constraint, etc.)\n")
	prompt.WriteString("3. Line number where the violation occurs (count from 1)\n")
	prompt.WriteString("4. A clear explanation of why it violates the rule\n")
	prompt.WriteString("5. A suggested fix (if applicable)\n\n")

	prompt.WriteString("Return your response as a JSON array with the following structure:\n")
	prompt.WriteString("[\n")
	prompt.WriteString("  {\n")
	prompt.WriteString("    \"object_name\": \"table_name.column_name\",\n")
	prompt.WriteString("    \"object_type\": \"column\",\n")
	prompt.WriteString("    \"line\": 15,\n")
	prompt.WriteString("    \"message\": \"Clear explanation of the violation\",\n")
	prompt.WriteString("    \"suggestion\": \"How to fix it\"\n")
	prompt.WriteString("  }\n")
	prompt.WriteString("]\n\n")
	prompt.WriteString("If no violations are found, return an empty array: []\n")
	prompt.WriteString("Return ONLY the JSON array, no additional text.")

	return prompt.String()
}

// BuildBatchPrompt creates a prompt for analyzing multiple rules at once
func (pb *PromptBuilder) BuildBatchPrompt(schema *models.Schema, rules []*models.Rule) string {
	var prompt strings.Builder

	prompt.WriteString("Analyze the following SQL schema against multiple lint rules:\n\n")
	
	// Schema information
	prompt.WriteString(fmt.Sprintf("SQL Dialect: %s\n\n", schema.Dialect))
	prompt.WriteString("SQL Schema:\n```sql\n")
	prompt.WriteString(schema.Raw)
	prompt.WriteString("\n```\n\n")

	// List all rules
	prompt.WriteString("Rules to check:\n\n")
	for i, rule := range rules {
		prompt.WriteString(fmt.Sprintf("%d. Rule ID: %s\n", i+1, rule.ID))
		prompt.WriteString(fmt.Sprintf("   Severity: %s\n", rule.Severity))
		prompt.WriteString(fmt.Sprintf("   Target: %s\n", rule.Target))
		prompt.WriteString(fmt.Sprintf("   Check: %s\n", rule.Check))
		prompt.WriteString(fmt.Sprintf("   Description: %s\n\n", rule.Description))
	}

	// Instructions
	prompt.WriteString("For each rule, check if the schema violates it. Return your response as a JSON object with rule IDs as keys:\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"rule-id-1\": [\n")
	prompt.WriteString("    {\n")
	prompt.WriteString("      \"object_name\": \"table_name.column_name\",\n")
	prompt.WriteString("      \"object_type\": \"column\",\n")
	prompt.WriteString("      \"line\": 15,\n")
	prompt.WriteString("      \"message\": \"Clear explanation of the violation\",\n")
	prompt.WriteString("      \"suggestion\": \"How to fix it\"\n")
	prompt.WriteString("    }\n")
	prompt.WriteString("  ],\n")
	prompt.WriteString("  \"rule-id-2\": []\n") // Empty array for no violations
	prompt.WriteString("}\n\n")
	prompt.WriteString("Return ONLY the JSON object, no additional text.")

	return prompt.String()
}

// ParseRuleResponse parses the AI response for a single rule
func (pb *PromptBuilder) ParseRuleResponse(response string, rule *models.Rule) ([]models.Issue, error) {
	// Clean the response
	response = strings.TrimSpace(response)
	
	// Try to extract JSON if wrapped in markdown code blocks
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json") + 7
		end := strings.LastIndex(response, "```")
		if start > 6 && end > start {
			response = response[start:end]
		}
	} else if strings.Contains(response, "```") {
		start := strings.Index(response, "```") + 3
		end := strings.LastIndex(response, "```")
		if start > 2 && end > start {
			response = response[start:end]
		}
	}

	// Parse JSON array
	var violations []struct {
		ObjectName string `json:"object_name"`
		ObjectType string `json:"object_type"`
		Line       int    `json:"line"`
		Message    string `json:"message"`
		Suggestion string `json:"suggestion"`
	}

	if err := json.Unmarshal([]byte(response), &violations); err != nil {
		// Try to extract JSON array from response
		startIdx := strings.Index(response, "[")
		endIdx := strings.LastIndex(response, "]")
		if startIdx >= 0 && endIdx > startIdx {
			jsonStr := response[startIdx : endIdx+1]
			if err := json.Unmarshal([]byte(jsonStr), &violations); err != nil {
				return nil, fmt.Errorf("failed to parse AI response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse AI response: %w", err)
		}
	}

	// Convert to Issue objects
	issues := make([]models.Issue, 0, len(violations))
	for _, v := range violations {
		issue := models.Issue{
			RuleID:     rule.ID,
			Severity:   rule.Severity,
			ObjectType: v.ObjectType,
			ObjectName: v.ObjectName,
			Message:    v.Message,
			Line:       v.Line,
			Column:     1, // AI typically doesn't provide column info
			Suggestion: v.Suggestion,
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// ParseBatchResponse parses the AI response for multiple rules
func (pb *PromptBuilder) ParseBatchResponse(response string, rules []*models.Rule) (map[string][]models.Issue, error) {
	// Clean the response
	response = strings.TrimSpace(response)
	
	// Try to extract JSON if wrapped in markdown code blocks
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json") + 7
		end := strings.LastIndex(response, "```")
		if start > 6 && end > start {
			response = response[start:end]
		}
	}

	// Parse JSON object
	var violations map[string][]struct {
		ObjectName string `json:"object_name"`
		ObjectType string `json:"object_type"`
		Line       int    `json:"line"`
		Message    string `json:"message"`
		Suggestion string `json:"suggestion"`
	}

	if err := json.Unmarshal([]byte(response), &violations); err != nil {
		return nil, fmt.Errorf("failed to parse batch AI response: %w", err)
	}

	// Convert to Issue objects
	result := make(map[string][]models.Issue)
	
	// Create a map of rule IDs to rules for quick lookup
	ruleMap := make(map[string]*models.Rule)
	for _, rule := range rules {
		ruleMap[rule.ID] = rule
	}

	for ruleID, ruleViolations := range violations {
		rule, exists := ruleMap[ruleID]
		if !exists {
			continue // Skip unknown rules
		}

		issues := make([]models.Issue, 0, len(ruleViolations))
		for _, v := range ruleViolations {
			issue := models.Issue{
				RuleID:     rule.ID,
				Severity:   rule.Severity,
				ObjectType: v.ObjectType,
				ObjectName: v.ObjectName,
				Message:    v.Message,
				Line:       v.Line,
				Column:     1,
				Suggestion: v.Suggestion,
			}
			issues = append(issues, issue)
		}
		result[ruleID] = issues
	}

	return result, nil
}