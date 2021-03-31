package modconfig

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/turbot/steampipe/constants"
)

type Query struct {
	Name        string
	Title       string `hcl:"title"`
	Description string `hcl:"description"`
	SQL         string `hcl:"sql"`
}

func (q *Query) String() string {
	return fmt.Sprintf(`
  -----
  Name: %s
  Title: %s
  Description: %s
  SQL: %s
`, q.Name, q.Title, q.Description, q.SQL)
}

func (q *Query) Equals(other *Query) bool {
	return q.Name == other.Name &&
		q.Title == other.Title &&
		q.Description == other.Description &&
		q.SQL == other.SQL
}

// factory function
func QueryFromFile(path string) (MappableResource, error) {
	q := &Query{}
	return q.InitialiseFromFile(path)
}

// implementation of MappableResource
func (q *Query) InitialiseFromFile(path string) (MappableResource, error) {
	// only valid for sql files
	if filepath.Ext(path) != constants.ExtensionSql {
		return nil, fmt.Errorf("Query.InitialiseFromFile must be called with .sql file only - got %s", path)
	}

	sqlBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	sql := string(sqlBytes)
	if sql == "" {
		return nil, fmt.Errorf("SQL file %s contains no query", path)
	}
	// get a sluggified version of the filename
	name, err := PseudoResourceNameFromPath(path)
	if err != nil {
		return nil, err
	}
	q.Name = name
	q.SQL = sql
	return q, nil
}
