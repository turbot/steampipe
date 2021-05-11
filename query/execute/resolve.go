package execute

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
)

// GetQueries retrieves queries from args
//
// For each arg check if it is a named query or a file, before falling back to treating it as sql
func GetQueries(args []string, workspace *workspace.Workspace) []string {
	var queries []string
	for _, arg := range args {
		query, _ := GetQueryFromArg(arg, workspace)
		if len(query) > 0 {
			queries = append(queries, query)
		}
	}
	return queries
}

// GetQueryFromArg attempts to resolve 'arg' to a query
//
// the second return value indicates whether the arg was resolved as a named query/SQL file
func GetQueryFromArg(arg string, workspace *workspace.Workspace) (string, bool) {
	// 1) is this a named query
	if namedQuery, ok := workspace.GetNamedQuery(arg); ok {
		return typeHelpers.SafeString(namedQuery.SQL), true
	}

	// 	2) is this a file
	fileQuery, fileExists, err := getQueryFromFile(arg)
	if fileExists {
		if err != nil {
			utils.ShowWarning(fmt.Sprintf("error opening file '%s': %v", arg, err))
			return "", false
		}
		if len(fileQuery) == 0 {
			utils.ShowWarning(fmt.Sprintf("file '%s' does not contain any data", arg))
			// (just return the empty string - it will be filtered above)
		}
		return fileQuery, true
	}

	// 3) just use the arg string as is and assume it is valid SQL
	return arg, false
}

func getQueryFromFile(filename string) (string, bool, error) {
	log.Println("[TRACE] getQueryFromFiles: ", filename)

	// get absolute filename
	path, err := filepath.Abs(filename)
	if err != nil {
		return "", false, nil
	}
	// does it exist?
	if _, err := os.Stat(path); err != nil {
		// if this gives any error, return not exist. we may get a not found or a path too long for example
		return "", false, nil
	}

	// read file
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "", true, err
	}

	return string(fileBytes), true, nil
}
