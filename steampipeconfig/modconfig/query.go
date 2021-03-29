package modconfig

import "fmt"

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
	// TODO
	return q, nil
}
