package db_common

import (
	"context"
	"github.com/turbot/steampipe/pkg/constants"
	"sort"
)

// GetDefaultSearchPath builds default search path from the connection schemas, book-ended with public and internal
func GetDefaultSearchPath(ctx context.Context, foreignSchemaNames []string) []string {
	// default to foreign schema names
	searchPath := foreignSchemaNames

	sort.Strings(searchPath)
	// add the 'public' schema as the first schema in the search_path. This makes it
	// easier for users to build and work with their own tables, and since it's normally
	// empty, doesn't make using steampipe tables any more difficult.
	searchPath = append([]string{"public"}, searchPath...)
	// add 'internal' schema as last schema in the search path
	searchPath = append(searchPath, constants.FunctionSchema)

	return searchPath
}
