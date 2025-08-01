package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Severity levels for lint issues
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Rule represents a lint rule
type Rule struct {
	ID          string            `json:"id"`
	Severity    Severity          `json:"severity"`
	Target      string            `json:"target"`
	Check       string            `json:"check"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags,omitempty"`
	GoodExample string            `json:"good_example,omitempty"`
	BadExample  string            `json:"bad_example,omitempty"`
}

// Issue represents a lint issue found in the schema
type Issue struct {
	RuleID     string   `json:"rule_id"`
	Severity   Severity `json:"severity"`
	ObjectType string   `json:"object_type"`
	ObjectName string   `json:"object_name"`
	Message    string   `json:"message"`
	Line       int      `json:"line"`
	Column     int      `json:"column"`
	Suggestion string   `json:"suggestion,omitempty"`
}

// LintResult represents the complete lint result
type LintResult struct {
	SchemaFile  string   `json:"schema_file"`
	Dialect     string   `json:"dialect"`
	TotalIssues int      `json:"total_issues"`
	Errors      int      `json:"errors"`
	Warnings    int      `json:"warnings"`
	Info        int      `json:"info"`
	Issues      []Issue  `json:"issues"`
}

// ToJSON converts the lint result to JSON format
func (r *LintResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToText converts the lint result to text format
func (r *LintResult) ToText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("SQL Schema Lint Results for %s\n", r.SchemaFile))
	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Dialect: %s\n", r.Dialect))
	sb.WriteString(fmt.Sprintf("Total Issues: %d (%d errors, %d warnings, %d info)\n\n", 
		r.TotalIssues, r.Errors, r.Warnings, r.Info))

	for _, issue := range r.Issues {
		severityUpper := strings.ToUpper(string(issue.Severity))
		sb.WriteString(fmt.Sprintf("%s: %s '%s' (line %d)\n", 
			severityUpper, issue.ObjectType, issue.ObjectName, issue.Line))
		sb.WriteString(fmt.Sprintf("  Rule: %s\n", issue.RuleID))
		sb.WriteString(fmt.Sprintf("  %s\n", issue.Message))
		if issue.Suggestion != "" {
			sb.WriteString(fmt.Sprintf("  Suggestion: %s\n", issue.Suggestion))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ToMarkdown converts the lint result to markdown format
func (r *LintResult) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# SQL Schema Lint Results\n\n"))
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n", r.SchemaFile))
	sb.WriteString(fmt.Sprintf("**Dialect:** %s\n", r.Dialect))
	sb.WriteString(fmt.Sprintf("**Total Issues:** %d\n\n", r.TotalIssues))

	if r.TotalIssues == 0 {
		sb.WriteString("✅ No issues found!\n")
		return sb.String()
	}

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- 🔴 **Errors:** %d\n", r.Errors))
	sb.WriteString(fmt.Sprintf("- 🟡 **Warnings:** %d\n", r.Warnings))
	sb.WriteString(fmt.Sprintf("- 🔵 **Info:** %d\n\n", r.Info))

	sb.WriteString("## Issues\n\n")

	// Group issues by severity
	var errors, warnings, infos []Issue
	for _, issue := range r.Issues {
		switch issue.Severity {
		case SeverityError:
			errors = append(errors, issue)
		case SeverityWarning:
			warnings = append(warnings, issue)
		case SeverityInfo:
			infos = append(infos, issue)
		}
	}

	if len(errors) > 0 {
		sb.WriteString("### 🔴 Errors\n\n")
		for _, issue := range errors {
			sb.WriteString(formatMarkdownIssue(issue))
		}
	}

	if len(warnings) > 0 {
		sb.WriteString("### 🟡 Warnings\n\n")
		for _, issue := range warnings {
			sb.WriteString(formatMarkdownIssue(issue))
		}
	}

	if len(infos) > 0 {
		sb.WriteString("### 🔵 Info\n\n")
		for _, issue := range infos {
			sb.WriteString(formatMarkdownIssue(issue))
		}
	}

	return sb.String()
}

func formatMarkdownIssue(issue Issue) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("#### %s `%s` (line %d)\n\n", 
		issue.ObjectType, issue.ObjectName, issue.Line))
	sb.WriteString(fmt.Sprintf("- **Rule:** %s\n", issue.RuleID))
	sb.WriteString(fmt.Sprintf("- **Message:** %s\n", issue.Message))
	if issue.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("- **Suggestion:** %s\n", issue.Suggestion))
	}
	sb.WriteString("\n")

	return sb.String()
}

// Schema represents a parsed SQL schema
type Schema struct {
	Dialect string              `json:"dialect"`
	Tables  map[string]*Table   `json:"tables"`
	Views   map[string]*View    `json:"views"`
	Raw     string              `json:"-"` // Original SQL content
}

// Table represents a SQL table
type Table struct {
	Name        string                 `json:"name"`
	Columns     map[string]*Column     `json:"columns"`
	PrimaryKey  *PrimaryKey            `json:"primary_key,omitempty"`
	ForeignKeys map[string]*ForeignKey `json:"foreign_keys,omitempty"`
	Indexes     map[string]*Index      `json:"indexes,omitempty"`
	Constraints map[string]*Constraint `json:"constraints,omitempty"`
	Comment     string                 `json:"comment,omitempty"`
	Line        int                    `json:"line"`
}

// Column represents a table column
type Column struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value,omitempty"`
	Comment      string `json:"comment,omitempty"`
	Line         int    `json:"line"`
}

// PrimaryKey represents a primary key constraint
type PrimaryKey struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Line    int      `json:"line"`
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	Name             string   `json:"name"`
	Columns          []string `json:"columns"`
	ReferencedTable  string   `json:"referenced_table"`
	ReferencedColumns []string `json:"referenced_columns"`
	OnDelete         string   `json:"on_delete,omitempty"`
	OnUpdate         string   `json:"on_update,omitempty"`
	Line             int      `json:"line"`
}

// Index represents a table index
type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Line    int      `json:"line"`
}

// Constraint represents a table constraint
type Constraint struct {
	Name       string `json:"name"`
	Type       string `json:"type"` // CHECK, UNIQUE, etc.
	Definition string `json:"definition"`
	Line       int    `json:"line"`
}

// View represents a SQL view
type View struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Line       int    `json:"line"`
}