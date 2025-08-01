package parser

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/rebeliceyang/schema-lint-mcp-server/pkg/models"
	"github.com/xwb1989/sqlparser"
)

type SQLParser struct {
	dialect string
}

func NewSQLParser(dialect string) *SQLParser {
	return &SQLParser{
		dialect: dialect,
	}
}

func (p *SQLParser) ParseFile(filePath string) (*models.Schema, error) {
	log.Printf("[SQLParser] Parsing SQL file: %s", filePath)
	
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[SQLParser] ERROR: Failed to open file: %v", err)
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[SQLParser] ERROR: Failed to read file: %v", err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	log.Printf("[SQLParser] File read successfully, size: %d bytes", len(content))
	return p.Parse(string(content))
}

func (p *SQLParser) Parse(sql string) (*models.Schema, error) {
	log.Printf("[SQLParser] Starting to parse SQL (length: %d, dialect: %s)", len(sql), p.dialect)
	
	schema := &models.Schema{
		Dialect: p.dialect,
		Tables:  make(map[string]*models.Table),
		Views:   make(map[string]*models.View),
		Raw:     sql,
	}

	// If dialect is not specified, try to auto-detect
	if p.dialect == "" {
		p.dialect = p.detectDialect(sql)
		schema.Dialect = p.dialect
		log.Printf("[SQLParser] Auto-detected dialect: %s", p.dialect)
	}

	// Split SQL into statements
	statements := p.splitStatements(sql)

	// Parse each statement
	for _, stmt := range statements {
		if err := p.parseStatement(schema, stmt); err != nil {
			// Continue parsing other statements even if one fails
			// Store error information for reporting
			continue
		}
	}

	return schema, nil
}

func (p *SQLParser) detectDialect(sql string) string {
	sqlLower := strings.ToLower(sql)

	// PostgreSQL specific features
	if strings.Contains(sqlLower, "serial") || 
		strings.Contains(sqlLower, "uuid_generate") ||
		strings.Contains(sqlLower, "jsonb") ||
		strings.Contains(sqlLower, "::") {
		return "postgres"
	}

	// MySQL specific features
	if strings.Contains(sqlLower, "auto_increment") ||
		strings.Contains(sqlLower, "engine=") ||
		strings.Contains(sqlLower, "charset=") {
		return "mysql"
	}

	// SQL Server specific features
	if strings.Contains(sqlLower, "identity(") ||
		strings.Contains(sqlLower, "[dbo]") ||
		strings.Contains(sqlLower, "nvarchar") {
		return "sqlserver"
	}

	// Oracle specific features
	if strings.Contains(sqlLower, "number(") ||
		strings.Contains(sqlLower, "varchar2") ||
		strings.Contains(sqlLower, "sequence") {
		return "oracle"
	}

	// Default to postgres
	return "postgres"
}

func (p *SQLParser) splitStatements(sql string) []statement {
	var statements []statement
	lines := strings.Split(sql, "\n")
	var currentStmt strings.Builder
	startLine := 1
	inStatement := false

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		if !inStatement {
			startLine = lineNum
			inStatement = true
		}

		currentStmt.WriteString(line)
		currentStmt.WriteString("\n")

		// Check if statement ends with semicolon
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, statement{
				sql:  currentStmt.String(),
				line: startLine,
			})
			currentStmt.Reset()
			inStatement = false
		}
	}

	// Add any remaining statement
	if currentStmt.Len() > 0 {
		statements = append(statements, statement{
			sql:  currentStmt.String(),
			line: startLine,
		})
	}

	return statements
}

type statement struct {
	sql  string
	line int
}

func (p *SQLParser) parseStatement(schema *models.Schema, stmt statement) error {
	sql := strings.TrimSpace(stmt.sql)
	sqlLower := strings.ToLower(sql)

	switch {
	case strings.HasPrefix(sqlLower, "create table"):
		return p.parseCreateTable(schema, sql, stmt.line)
	case strings.HasPrefix(sqlLower, "create view"):
		return p.parseCreateView(schema, sql, stmt.line)
	case strings.HasPrefix(sqlLower, "alter table"):
		return p.parseAlterTable(schema, sql, stmt.line)
	case strings.HasPrefix(sqlLower, "create index"):
		return p.parseCreateIndex(schema, sql, stmt.line)
	}

	return nil
}

func (p *SQLParser) parseCreateTable(schema *models.Schema, sql string, line int) error {
	// Try using sqlparser for standard SQL
	stmt, err := sqlparser.Parse(sql)
	if err == nil {
		if createTable, ok := stmt.(*sqlparser.DDL); ok && createTable.Action == "create" {
			return p.parseCreateTableAST(schema, createTable, line)
		}
	}

	// Fallback to regex parsing for dialect-specific syntax
	return p.parseCreateTableRegex(schema, sql, line)
}

func (p *SQLParser) parseCreateTableAST(schema *models.Schema, stmt *sqlparser.DDL, line int) error {
	// For now, just use regex parsing since sqlparser DDL structure is different
	// In a real implementation, we would parse the DDL structure properly
	return fmt.Errorf("AST parsing not implemented, falling back to regex")
}

func (p *SQLParser) parseCreateTableRegex(schema *models.Schema, sql string, line int) error {
	// Extract table name
	tableNameRe := regexp.MustCompile(`(?i)create\s+table\s+(?:if\s+not\s+exists\s+)?(["\w` + "`" + `]+)\s*\(`)
	matches := tableNameRe.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse table name")
	}

	tableName := p.cleanIdentifier(matches[1])
	table := &models.Table{
		Name:        tableName,
		Columns:     make(map[string]*models.Column),
		ForeignKeys: make(map[string]*models.ForeignKey),
		Indexes:     make(map[string]*models.Index),
		Constraints: make(map[string]*models.Constraint),
		Line:        line,
	}

	// Extract table definition
	start := strings.Index(sql, "(")
	end := strings.LastIndex(sql, ")")
	if start == -1 || end == -1 || start >= end {
		return fmt.Errorf("invalid table definition")
	}

	definition := sql[start+1 : end]
	
	// Parse columns and constraints
	elements := p.splitTableElements(definition)
	for i, element := range elements {
		element = strings.TrimSpace(element)
		elementLower := strings.ToLower(element)

		switch {
		case strings.HasPrefix(elementLower, "primary key"):
			p.parsePrimaryKey(table, element, line+i)
		case strings.HasPrefix(elementLower, "foreign key"):
			p.parseForeignKey(table, element, line+i)
		case strings.HasPrefix(elementLower, "constraint"):
			p.parseConstraint(table, element, line+i)
		case strings.HasPrefix(elementLower, "index") || strings.HasPrefix(elementLower, "key"):
			p.parseInlineIndex(table, element, line+i)
		default:
			// Parse as column
			p.parseColumn(table, element, line+i)
		}
	}

	schema.Tables[tableName] = table
	return nil
}

func (p *SQLParser) splitTableElements(definition string) []string {
	var elements []string
	var current strings.Builder
	parenDepth := 0
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range definition {
		switch r {
		case '(':
			if !inQuotes {
				parenDepth++
			}
		case ')':
			if !inQuotes {
				parenDepth--
			}
		case '"', '\'', '`':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
			}
		case ',':
			if parenDepth == 0 && !inQuotes {
				elements = append(elements, current.String())
				current.Reset()
				continue
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		elements = append(elements, current.String())
	}

	return elements
}

func (p *SQLParser) parseColumn(table *models.Table, columnDef string, line int) {
	// Basic column parsing regex
	columnRe := regexp.MustCompile(`^(["\w` + "`" + `]+)\s+([A-Za-z0-9_().,\s]+)(.*)$`)
	matches := columnRe.FindStringSubmatch(columnDef)
	if len(matches) < 3 {
		return
	}

	column := &models.Column{
		Name:     p.cleanIdentifier(matches[1]),
		Type:     strings.TrimSpace(matches[2]),
		Nullable: true,
		Line:     line,
	}

	// Parse column modifiers
	modifiers := strings.ToLower(matches[3])
	if strings.Contains(modifiers, "not null") {
		column.Nullable = false
	}
	if strings.Contains(modifiers, "primary key") {
		// Handle inline primary key
		if table.PrimaryKey == nil {
			table.PrimaryKey = &models.PrimaryKey{
				Name:    fmt.Sprintf("pk_%s", table.Name),
				Columns: []string{column.Name},
				Line:    line,
			}
		}
	}

	// Extract default value
	defaultRe := regexp.MustCompile(`(?i)default\s+([^,]+)`)
	if defaultMatches := defaultRe.FindStringSubmatch(modifiers); len(defaultMatches) > 1 {
		column.DefaultValue = strings.TrimSpace(defaultMatches[1])
	}

	table.Columns[column.Name] = column
}

func (p *SQLParser) parsePrimaryKey(table *models.Table, constraint string, line int) {
	// Extract columns
	columnsRe := regexp.MustCompile(`(?i)primary\s+key\s*\(([^)]+)\)`)
	matches := columnsRe.FindStringSubmatch(constraint)
	if len(matches) < 2 {
		return
	}

	columns := p.parseColumnList(matches[1])
	table.PrimaryKey = &models.PrimaryKey{
		Name:    fmt.Sprintf("pk_%s", table.Name),
		Columns: columns,
		Line:    line,
	}
}

func (p *SQLParser) parseForeignKey(table *models.Table, constraint string, line int) {
	// Parse foreign key constraint
	fkRe := regexp.MustCompile(`(?i)(?:constraint\s+(["\w` + "`" + `]+)\s+)?foreign\s+key\s*\(([^)]+)\)\s+references\s+(["\w` + "`" + `.]+)\s*\(([^)]+)\)(.*)`)
	matches := fkRe.FindStringSubmatch(constraint)
	if len(matches) < 5 {
		return
	}

	fkName := p.cleanIdentifier(matches[1])
	if fkName == "" {
		fkName = fmt.Sprintf("fk_%s_%d", table.Name, len(table.ForeignKeys)+1)
	}

	fk := &models.ForeignKey{
		Name:              fkName,
		Columns:           p.parseColumnList(matches[2]),
		ReferencedTable:   p.cleanIdentifier(matches[3]),
		ReferencedColumns: p.parseColumnList(matches[4]),
		Line:              line,
	}

	// Parse ON DELETE/UPDATE actions
	actions := strings.ToLower(matches[5])
	if onDelete := regexp.MustCompile(`on\s+delete\s+(\w+)`).FindStringSubmatch(actions); len(onDelete) > 1 {
		fk.OnDelete = strings.ToUpper(onDelete[1])
	}
	if onUpdate := regexp.MustCompile(`on\s+update\s+(\w+)`).FindStringSubmatch(actions); len(onUpdate) > 1 {
		fk.OnUpdate = strings.ToUpper(onUpdate[1])
	}

	table.ForeignKeys[fkName] = fk
}

func (p *SQLParser) parseConstraint(table *models.Table, constraint string, line int) {
	// Parse named constraints
	constraintRe := regexp.MustCompile(`(?i)constraint\s+(["\w` + "`" + `]+)\s+(.+)`)
	matches := constraintRe.FindStringSubmatch(constraint)
	if len(matches) < 3 {
		return
	}

	name := p.cleanIdentifier(matches[1])
	definition := strings.TrimSpace(matches[2])
	definitionLower := strings.ToLower(definition)

	// Determine constraint type
	var constraintType string
	switch {
	case strings.HasPrefix(definitionLower, "check"):
		constraintType = "CHECK"
	case strings.HasPrefix(definitionLower, "unique"):
		constraintType = "UNIQUE"
	default:
		return
	}

	table.Constraints[name] = &models.Constraint{
		Name:       name,
		Type:       constraintType,
		Definition: definition,
		Line:       line,
	}
}

func (p *SQLParser) parseInlineIndex(table *models.Table, indexDef string, line int) {
	// Parse inline index definition
	indexRe := regexp.MustCompile(`(?i)(?:unique\s+)?(?:index|key)\s+(["\w` + "`" + `]+)?\s*\(([^)]+)\)`)
	matches := indexRe.FindStringSubmatch(indexDef)
	if len(matches) < 3 {
		return
	}

	indexName := p.cleanIdentifier(matches[1])
	if indexName == "" {
		indexName = fmt.Sprintf("idx_%s_%d", table.Name, len(table.Indexes)+1)
	}

	index := &models.Index{
		Name:    indexName,
		Columns: p.parseColumnList(matches[2]),
		Unique:  strings.Contains(strings.ToLower(indexDef), "unique"),
		Line:    line,
	}

	table.Indexes[indexName] = index
}

func (p *SQLParser) parseCreateView(schema *models.Schema, sql string, line int) error {
	// Extract view name
	viewNameRe := regexp.MustCompile(`(?i)create\s+(?:or\s+replace\s+)?view\s+(["\w` + "`" + `.]+)\s+as`)
	matches := viewNameRe.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse view name")
	}

	viewName := p.cleanIdentifier(matches[1])
	view := &models.View{
		Name:       viewName,
		Definition: sql,
		Line:       line,
	}

	schema.Views[viewName] = view
	return nil
}

func (p *SQLParser) parseAlterTable(schema *models.Schema, sql string, line int) error {
	// Extract table name
	tableNameRe := regexp.MustCompile(`(?i)alter\s+table\s+(["\w` + "`" + `.]+)`)
	matches := tableNameRe.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse table name in ALTER TABLE")
	}

	tableName := p.cleanIdentifier(matches[1])
	table, exists := schema.Tables[tableName]
	if !exists {
		// Create table if it doesn't exist (might be defined in another file)
		table = &models.Table{
			Name:        tableName,
			Columns:     make(map[string]*models.Column),
			ForeignKeys: make(map[string]*models.ForeignKey),
			Indexes:     make(map[string]*models.Index),
			Constraints: make(map[string]*models.Constraint),
			Line:        line,
		}
		schema.Tables[tableName] = table
	}

	// Parse ALTER TABLE operations
	sqlLower := strings.ToLower(sql)
	switch {
	case strings.Contains(sqlLower, "add column"):
		// Parse ADD COLUMN
		addColumnRe := regexp.MustCompile(`(?i)add\s+column\s+(["\w` + "`" + `]+)\s+(.+)`)
		if matches := addColumnRe.FindStringSubmatch(sql); len(matches) > 2 {
			p.parseColumn(table, matches[1]+" "+matches[2], line)
		}
	case strings.Contains(sqlLower, "add constraint"):
		// Parse ADD CONSTRAINT
		constraintRe := regexp.MustCompile(`(?i)add\s+constraint\s+(.+)`)
		if matches := constraintRe.FindStringSubmatch(sql); len(matches) > 1 {
			p.parseConstraint(table, "constraint "+matches[1], line)
		}
	case strings.Contains(sqlLower, "add foreign key"):
		// Parse ADD FOREIGN KEY
		fkRe := regexp.MustCompile(`(?i)add\s+foreign\s+key\s+(.+)`)
		if matches := fkRe.FindStringSubmatch(sql); len(matches) > 1 {
			p.parseForeignKey(table, "foreign key "+matches[1], line)
		}
	}

	return nil
}

func (p *SQLParser) parseCreateIndex(schema *models.Schema, sql string, line int) error {
	// Parse CREATE INDEX statement
	indexRe := regexp.MustCompile(`(?i)create\s+(?:unique\s+)?index\s+(?:if\s+not\s+exists\s+)?(["\w` + "`" + `]+)\s+on\s+(["\w` + "`" + `.]+)\s*\(([^)]+)\)`)
	matches := indexRe.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return fmt.Errorf("failed to parse CREATE INDEX")
	}

	indexName := p.cleanIdentifier(matches[1])
	tableName := p.cleanIdentifier(matches[2])
	columns := p.parseColumnList(matches[3])

	table, exists := schema.Tables[tableName]
	if !exists {
		// Create table if it doesn't exist
		table = &models.Table{
			Name:        tableName,
			Columns:     make(map[string]*models.Column),
			ForeignKeys: make(map[string]*models.ForeignKey),
			Indexes:     make(map[string]*models.Index),
			Constraints: make(map[string]*models.Constraint),
			Line:        line,
		}
		schema.Tables[tableName] = table
	}

	index := &models.Index{
		Name:    indexName,
		Columns: columns,
		Unique:  strings.Contains(strings.ToLower(sql), "unique"),
		Line:    line,
	}

	table.Indexes[indexName] = index
	return nil
}

func (p *SQLParser) parseColumnList(columnList string) []string {
	var columns []string
	parts := strings.Split(columnList, ",")
	for _, part := range parts {
		column := p.cleanIdentifier(strings.TrimSpace(part))
		if column != "" {
			columns = append(columns, column)
		}
	}
	return columns
}

func (p *SQLParser) cleanIdentifier(identifier string) string {
	// Remove quotes and backticks
	identifier = strings.Trim(identifier, " \t\n\r")
	identifier = strings.Trim(identifier, "\"'`")
	
	// Handle schema.table notation
	parts := strings.Split(identifier, ".")
	if len(parts) > 1 {
		// Return just the table name for now
		return strings.Trim(parts[len(parts)-1], "\"'`")
	}
	
	return identifier
}