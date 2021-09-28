package workspace

import "github.com/turbot/steampipe/steampipeconfig/modconfig"

type SessionDataSource struct {
	PreparedStatementSource, IntrospectionTableSource *modconfig.WorkspaceResourceMaps
}

// NewSessionDataSource creates a new SessionDataSource object
// it defaults to using the same source for prepared statemntrs and introspection tables
func NewSessionDataSource(source *modconfig.WorkspaceResourceMaps) *SessionDataSource {
	return &SessionDataSource{
		IntrospectionTableSource: source,
		PreparedStatementSource:  source,
	}

}
