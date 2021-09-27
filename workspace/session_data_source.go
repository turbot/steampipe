package workspace

import "github.com/turbot/steampipe/steampipeconfig/modconfig"

type SessionDataSource struct {
	preparedStatementSource, introspectionTableSource *modconfig.WorkspaceResourceMaps
}

// NewSessionDataSource creates a new SessionDataSource object
// if a single parameter is poassed, this map is used for both prepared statements and introspection tables
// if a second parameter is passed, it will be a minimal set of resources for which we need to create prepared statements
// this will be populated for batch mode querying
func NewSessionDataSource(items ...*modconfig.WorkspaceResourceMaps) *SessionDataSource {
	if len(items) == 0 {
		panic("NewSessionStateSource called with no parameters")
	}
	if len(items) > 2 {
		panic("NewSessionStateSource called with more than 2 parameters")
	}
	// default to initialising introspectionTableSource AND preparedStatementSource from the first param,
	// which is expected to be the full map of workspace resources
	res := &SessionDataSource{
		introspectionTableSource: items[0],
		preparedStatementSource:  items[0],
	}
	// is the preparedStatementSource explicitly provided?
	if len(items) == 2 {
		res.preparedStatementSource = items[1]
	}
	return res

}
