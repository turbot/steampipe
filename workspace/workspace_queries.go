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
func (w *Workspace) GetQueriesFromArgs(args []string) ([]string, *modconfig.WorkspaceResourceMaps, error) {
	utils.LogTime("execute.GetQueriesFromArgs start")
	defer utils.LogTime("execute.GetQueriesFromArgs end")

	var queries []string
	// build map of prepared statement providers
	var resourceMap = modconfig.GetWorkspaceResourceMapForMod(w.Mod)
	for _, arg := range args {
		query, queryProvider, err := w.ResolveQueryAndArgs(arg)
		if err != nil {
			return nil, nil, err
		}
		if len(query) > 0 {
			queries = append(queries, query)
			resourceMap.AddQueryProvider(queryProvider)
		}
	}
	return queries, resourceMap, nil
}

// ResolveQueryAndArgs attempts to resolve 'arg' to a query and query args
func (w *Workspace) ResolveQueryAndArgs(sqlString string) (string, modconfig.QueryProvider, error) {
	var args = &modconfig.QueryArgs{}

	var err error

	// if this looks like a named query or named control invocation, parse the sql string for arguments
	if isNamedQueryOrControl(sqlString) {
		sqlString, args, err = parse.ParsePreparedStatementInvocation(sqlString)
		if err != nil {
			return "", nil, err
		}
	}

	return w.resolveQuery(sqlString, args)
}

// ResolveQuery resolves the query for the given QueryProvider
func (w *Workspace) ResolveQuery(queryProvider modconfig.QueryProvider, args *modconfig.QueryArgs) (string, error) {
	args, err := w.resolveQueryProviderArgs(queryProvider, args)
	if err != nil {
		return "", err
	}

	log.Printf("[TRACE] ResolveQuery for %s", queryProvider.Name())

	query := queryProvider.GetQuery()
	sql := queryProvider.GetSQL()
	params := queryProvider.GetParams()
	// verify we have either SQL or a Query defined
	if sql == nil && query == nil {
		// this should never happen as we should catch it in the parsing stage
		return "", fmt.Errorf("%s must define either a 'sql' property or a 'query' property", queryProvider.Name())
	}

	// determine the source for the query
	// - this will either be the control itself or any named query the control refers to
	// either via its SQL property (passing a query name) or Query property (using a reference to a query object)

	// if a query is provided, us that to resolve the sql
	if query != nil {
		return w.resolveNamedQuery(query, args)
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
			return w.resolveNamedQuery(namedQuery, args)
		}
		// so the control sql is NOT a named query
		// if there are NO params, use the control SQL as is
		if len(params) == 0 {
			return queryProviderSQL, nil
		}
		// so the control sql is NOT a named query
		// if there are NO params, use the control SQL as is
		if len(params) == 0 {
			return queryProviderSQL, nil
		}
	}

	// so the control defines SQL and has params - it is a prepared statement
	return queryProvider.GetPreparedStatementExecuteSQL(args)
}

func (w *Workspace) resolveQuery(sqlString string, args *modconfig.QueryArgs) (string, modconfig.QueryProvider, error) {
	// query or control providing the named query

	log.Printf("[TRACE] resolveQuery %s args %s", sqlString, args)
	// 1) check if this is a control
	if control, ok := w.GetControl(sqlString); ok {

		log.Printf("[TRACE] query string is a control: %s", control.FullName)

		// copy control SQL into query and continue resolution
		var err error
		sqlString, err = w.ResolveQuery(control, args)
		if err != nil {
			return "", nil, err
		}
		log.Printf("[TRACE] resolved control query: %s", sqlString)
		return sqlString, control, nil
	}

	// 2) is this a named query
	if namedQuery, ok := w.GetQuery(sqlString); ok {
		sql, err := w.resolveNamedQuery(namedQuery, args)
		if err != nil {
			return "", nil, err
		}
		return sql, namedQuery, nil
	}

	// 	3) is this a file
	fileQuery, fileExists, err := w.getQueryFromFile(sqlString)
	if fileExists {
		if err != nil {
			return "", nil, fmt.Errorf("ResolveQueryAndArgs failed: error opening file '%s': %v", sqlString, err)
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

func (w *Workspace) resolveNamedQuery(namedQuery *modconfig.Query, args *modconfig.QueryArgs) (string, error) {
	/// if there are no params, just return the sql
	if len(namedQuery.Params) == 0 {
		return typehelpers.SafeString(namedQuery.SQL), nil
	}

	// so there are params - this will be a prepared statement
	sql, err := modconfig.GetPreparedStatementExecuteSQL(namedQuery, args)
	if err != nil {
		return "", err
	}
	return sql, nil
}

func (w *Workspace) resolveQueryProviderArgs(queryProvider modconfig.QueryProvider, args *modconfig.QueryArgs) (*modconfig.QueryArgs, error) {
	// if no args were provided,  set args to control args (which may also be nil!)
	if args == nil || args.Empty() {
		log.Printf("[TRACE] using control args: %s", queryProvider.GetArgs())
		return queryProvider.GetArgs(), nil
	}
	// so command line args were provided
	// check if the control supports them (it will NOT is it specifies a 'query' property)
	if queryProvider.GetQuery() != nil {
		return nil, fmt.Errorf("%s defines a query property and so does not support command line arguments", queryProvider.Name())
	}
	log.Printf("[TRACE] using command line args: %s", args)

	// so the control defines SQL and has params - it is a prepared statement
	return args, nil
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
