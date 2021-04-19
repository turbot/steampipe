package modconfig

import (
	"fmt"
	"reflect"

	"github.com/turbot/go-kit/types"
)

type Control struct {
	Name        *string
	Title       *string           `hcl:"title"`
	Description *string           `hcl:"description"`
	Tags        map[string]string `hcl:"tags"`
	SQL         *string           `hcl:"sql"`
	DocLink     *string           `hcl:"doc_link"`
}

func (c *Control) String() string {
	return fmt.Sprintf(`
  -----
  Name: %s
  Title: %s
  Description: %s
  SQL: %s
`, types.SafeString(c.Name), types.SafeString(c.Title), types.SafeString(c.Description), types.SafeString(c.SQL))
}

func (c *Control) Equals(other *Control) bool {
	return types.SafeString(c.Name) == types.SafeString(other.Name) &&
		types.SafeString(c.Title) == types.SafeString(other.Title) &&
		types.SafeString(c.Description) == types.SafeString(other.Description) &&
		types.SafeString(c.SQL) == types.SafeString(other.SQL) &&
		types.SafeString(c.DocLink) == types.SafeString(other.DocLink) &&
		reflect.DeepEqual(c.Tags, other.Tags)
}
