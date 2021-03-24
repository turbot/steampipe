package modconfig

import "fmt"

type Query struct {
	Name        string `hcl:"name,label"`
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
