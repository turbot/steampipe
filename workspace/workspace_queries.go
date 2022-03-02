package workspace

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/utils"
)

// GetQueriesFromArgs retrieves queries from args
//
// For each arg check if it is a named query or a file, before falling back to treating it as sql
func (w *Workspace) GetQueriesFromArgs(args []string) ([]string, *modconfig.ModResources, error) {
	utils.LogTime("execute.GetQueriesFromArgs start")
	defer utils.LogTime("execute.GetQueriesFromArgs end")

	var queries []string
	var queryProviders []modconfig.QueryProvider
	// build map of just the required prepared statement providers
	for _, arg := range args {
		query, queryProvider, err := w.ResolveQueryAndArgsFromSQLString(arg)
		if err != nil {
			return nil, nil, err
		}
		if len(query) > 0 {
			queries = append(queries, query)
			queryProviders = append(queryProviders, queryProvider)

		}
	}
	var preparedStatementSource *modconfig.ModResources
	if len(queries) > 0 {
		preparedStatementSource = modconfig.CreateWorkspaceResourceMapForQueries(queryProviders, w.Mod)
	}
	return queries, preparedStatementSource, nil
}

// ResolveQueryAndArgsFromSQLString attempts to resolve 'arg' to a query and query args
func (w *Workspace) ResolveQueryAndArgsFromSQLString(sqlString string) (string, modconfig.QueryProvider, error) {
	var args = &modconfig.QueryArgs{}

	var err error

	// if this looks like a named query or named control invocation, parse the sql string for arguments
	if isNamedQueryOrControl(sqlString) {
		sqlString, args, err = parse.ParsePreparedStatementInvocation(sqlString)
		if err != nil {
			return "", nil, err
		}
	}
	// query or control providing the named query

	log.Printf("[TRACE] resolveQuery %s args %s", sqlString, args)
	// 1) check if this is a control
	if control, ok := w.GetControl(sqlString); ok {

		log.Printf("[TRACE] query string is a control: %s", control.FullName)

		// copy control SQL into query and continue resolution
		var err error
		sqlString, err = w.ResolveQueryFromQueryProvider(control, args)
		if err != nil {
			return "", nil, err
		}
		log.Printf("[TRACE] resolved control query: %s", sqlString)
		return sqlString, control, nil
	}

	// 2) is this a named query
	if namedQuery, ok := w.GetQuery(sqlString); ok {
		sql, err := w.ResolveQueryFromQueryProvider(namedQuery, args)
		if err != nil {
			return "", nil, err
		}
		return sql, namedQuery, nil
	}

	// 	3) is this a file
	fileQuery, fileExists, err := w.getQueryFromFile(sqlString)
	if fileExists {
		if err != nil {
			return "", nil, fmt.Errorf("ResolveQueryAndArgsFromSQLString failed: error opening file '%s': %v", sqlString, err)
		}
		if len(fileQuery) == 0 {
			utils.ShowWarning(fmt.Sprintf("file '%s' does not contain any data", sqlString))
			// (just return the empty string - it will be filtered above)
		}
		return fileQuery, nil, nil
	}

	// 4) so we have not managed to resolve this - if it looks like a named query or control, return an error
	if isNamedQueryOrControl(sqlString) {
		return "", nil, fmt.Errorf("'%s' not found in workspace", sqlString)
	}

	// 5) just use the query string as is and assume it is valid SQL
	return sqlString, nil, nil
}

// ResolveQueryFromQueryProvider resolves the query for the given QueryProvider
func (w *Workspace) ResolveQueryFromQueryProvider(queryProvider modconfig.QueryProvider, runtimeArgs *modconfig.QueryArgs) (string, error) {
	log.Printf("[TRACE] ResolveQueryFromQueryProvider for %s", queryProvider.Name())

	// verify the resource has qa query or sql, if required
	err := queryProvider.VerifyQuery(queryProvider)
	if err != nil {
		return "", err
	}

	query := queryProvider.GetQuery()
	sql := queryProvider.GetSQL()
	params := queryProvider.GetParams()

	// merge the base args with the runtime args
	runtimeArgs, err = modconfig.MergeArgs(queryProvider, runtimeArgs)
	if err != nil {
		return "", err
	}

	// determine the source for the query
	// - this will either be the control itself or any named query the control refers to
	// either via its SQL proper ty (passing a query name) or Query property (using a reference to a query object)

	// if a query is provided, use that to resolve the sql
	if query != nil {
		return w.ResolveQueryFromQueryProvider(query, runtimeArgs)
	}

	// if the control has SQL set, use that
	if sql != nil {
		queryProviderSQL := typehelpers.SafeString(sql)
		log.Printf("[TRACE] control defines inline SQL")

		// if the control SQL refers to a named query, this is the same as if the control 'Query' property is set
		if namedQuery, ok := w.GetQuery(queryProviderSQL); ok {
			// in this case, it is NOT valid for the control to define its own Param definitions
			if params != nil {
				return "", fmt.Errorf("%s has an 'SQL' property which refers to %s, so it cannot define 'param' blocks", queryProvider.Name(), namedQuery.FullName)
			}
			return w.ResolveQueryFromQueryProvider(namedQuery, runtimeArgs)
		}

		// so the control sql is NOT a named query

		// determine whether there are any params - there may either be param defs, OR positional args
		// if there are NO params OR list args, use the control SQL as is
		if !queryProvider.IsParameterised(runtimeArgs, params) {
			return queryProviderSQL, nil
		}
	}

	// so the control defines SQL and has params - it is a prepared statement
	return queryProvider.GetPreparedStatementExecuteSQL(runtimeArgs)
}

func (w *Workspace) getQueryFromFile(filename string) (string, bool, error) {
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

// does this resource name look like a control or query
func isNamedQueryOrControl(name string) bool {
	parsedResourceName, err := modconfig.ParseResourceName(name)
	return err == nil && (parsedResourceName.ItemType == "query" || parsedResourceName.ItemType == "control")
}
