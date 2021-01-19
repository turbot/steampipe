package schema

import (
	"regexp"
	"sort"
	"strings"
)

// Metadata :: struct to represent the schema of the database
type Metadata struct {
	// map {schemaname, {map tablename -> tableschema}}
	Schemas map[string]map[string]TableSchema
}

// TableSchema :: contains the details of a single table in the schema
type TableSchema struct {
	// map {columnName -> columnschema}
	Columns     map[string]ColumnSchema
	Name        string
	Schema      string
	Description string
}

// ColumnSchema :: contains the details of a single column in a table
type ColumnSchema struct {
	ID          string
	Name        string
	NotNull     bool
	Type        string
	Default     string
	Description string
}

// GetSchemas :: returns all foreign schema names
func (m *Metadata) GetSchemas() []string {
	schemas := []string{}
	for schema := range m.Schemas {
		schemas = append(schemas, schema)
	}
	sort.Strings(schemas)
	return schemas
}

// GetTablesInSchema :: returns all foreign tables in a given foreign schema
func (m *Metadata) GetTablesInSchema(schemaName string) []string {
	tables := []string{}
	for table := range m.Schemas[schemaName] {
		tables = append(tables, table)
	}
	return tables
}

// IsSchemaNameValid :: verifies that the given string is a valid pgsql schema name
func IsSchemaNameValid(name string) bool {

	// start with the basics

	// cannot be blank
	if len(strings.TrimSpace(name)) == 0 {
		return false
	}

	// there should not be whitespaces or dashes
	if strings.Contains(name, " ") || strings.Contains(name, "-") {
		return false
	}

	// cannot start with `pg_`
	if strings.HasPrefix(name, "pg_") {
		return false
	}

	// as per https://www.postgresql.org/docs/9.2/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS
	// not allowing $ sign, since it is not allowed in standard sql
	regex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)

	if !regex.MatchString(name) {
		return false
	}

	// let's limit the length to 63
	if len(name) > 63 {
		return false
	}

	return true
}
