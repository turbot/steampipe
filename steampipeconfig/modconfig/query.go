package modconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Query struct {
	Name        string
	Title       string `hcl:"title"`
	Description string `hcl:"description"`
	SQL         string `hcl:"sql"`
}

func (q *Query) String() string {
	return fmt.Sprintf(`  -----
  Name: %s
  Title: %s
  Description: %s
  SQL: %s`, q.Name, q.Title, q.Description, q.SQL)
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
	if filepath.Ext(path) != ".sql" {
		return nil, fmt.Errorf("Query.InitialiseFromFile must be called with .sql file only - got %s", path)
	}

	sql, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// get a sluggified version of the filename
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory when converting sql files to query resources: %v", err)
	}
	// get relative path of file
	relativePath := filepath.Rel(wd, path)
	// now slugify this
	q.SQL = string(sql)
	return q, nil
}
