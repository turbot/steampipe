package workspace

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type SessionDataSource struct {
	PreparedStatementSource  func() *modconfig.WorkspaceResourceMaps
	IntrospectionTableSource func() *modconfig.WorkspaceResourceMaps
}

func NewSessionDataSource(w *Workspace, preparedStatementSource *modconfig.WorkspaceResourceMaps) *SessionDataSource {
	res := &SessionDataSource{
		IntrospectionTableSource: func() *modconfig.WorkspaceResourceMaps {
			return w.GetResourceMaps()
		},
		PreparedStatementSource: func() *modconfig.WorkspaceResourceMaps {
			return w.GetResourceMaps()
		},
	}
	if preparedStatementSource != nil && !preparedStatementSource.Empty() {
		res.PreparedStatementSource = func() *modconfig.WorkspaceResourceMaps {
			return preparedStatementSource
		}
	}
	return res
}
