package schema

import (
	"regexp"
	"sort"
	"strings"
)

func NewMetadata() *Metadata {
	return &Metadata{
		Schemas:             map[string]map[string]TableSchema{},
		SearchPath:          []string{},
		TemporarySchemaName: "", // don't need this, adding for completeness
	}
}

// Metadata is a struct to represent the schema of the database
type Metadata struct {
	// map {schemaname, {map {tablename -> tableschema}}
	Schemas map[string]map[string]TableSchema
	// the search path that is set in the backend
	SearchPath []string
	// the name of the temporary schema
	TemporarySchemaName string
}

// TableSchema contains the details of a single table in the schema
type TableSchema struct {
	// map {columnName -> columnschema}
	Columns     map[string]ColumnSchema
	Name        string
	Schema      string
	Description string
}

// ColumnSchema contains the details of a single column in a table
type ColumnSchema struct {
	ID          string
	Name        string
	NotNull     bool
	Type        string
	Default     string
	Description string
}

// GetSchemas returns all foreign schema names
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
func IsSchemaNameValid(name string) (bool, string) {
	var message string

	// start with the basics

	// cannot be blank
	if len(strings.TrimSpace(name)) == 0 {
		message = "Schema name cannot be blank."
		return false, message
	}

	// there should not be whitespaces or dashes
	if strings.Contains(name, " ") || strings.Contains(name, "-") {
		message = "Schema name should not contain whitespaces or dashes."
		return false, message
	}

	// cannot start with `pg_`
	if strings.HasPrefix(name, "pg_") {
		message = "Schema name should not start with `pg_`"
		return false, message
	}

	// as per https://www.postgresql.org/docs/9.2/sql-syntax-lexical.html#SQL-SYNTAX-IDENTIFIERS
	// not allowing $ sign, since it is not allowed in standard sql
	regex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)

	if !regex.MatchString(name) {
		message = "Schema name string contains invalid pattern."
		return false, message
	}

	// let's limit the length to 63
	if len(name) > 63 {
		message = "Schema name length should not exceed 63 characters."
		return false, message
	}

	return true, message
}
