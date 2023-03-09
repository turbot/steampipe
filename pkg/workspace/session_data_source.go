package workspace

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type SessionDataSource struct {
	PreparedStatementSource  func() *modconfig.ResourceMaps
	IntrospectionTableSource func() *modconfig.ResourceMaps
}

// NewSessionDataSource uses the workspace and (optionally) a separate the prepared statemeot source
// and returns a SessionDataSource
// NOTE: preparedStatementSource is only set if specific queries have ben passed to the query command
// it allows us to only create the prepared statements me need
func NewSessionDataSource(w *Workspace, preparedStatementSource *modconfig.ResourceMaps) *SessionDataSource {
	res := &SessionDataSource{
		IntrospectionTableSource: func() *modconfig.ResourceMaps {
			return w.GetResourceMaps()
		},
		PreparedStatementSource: func() *modconfig.ResourceMaps {
			return w.GetResourceMaps()
		},
	}
	if preparedStatementSource != nil {
		res.PreparedStatementSource = func() *modconfig.ResourceMaps {
			return preparedStatementSource
		}
	}
	return res
}
