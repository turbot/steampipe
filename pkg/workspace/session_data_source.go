package workspace

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type SessionDataSource struct {
	PreparedStatementSource  func() *modconfig.ModResources
	IntrospectionTableSource func() *modconfig.ModResources
}

// NewSessionDataSource uses the workspace and (optionally) a separate the prepared statemeot source
// and returns a SessionDataSource
// NOTE: preparedStatementSource is only set if specific queries have ben passed to the query command
// it allows us to only create the prepared statements me need
func NewSessionDataSource(w *Workspace, preparedStatementSource *modconfig.ModResources) *SessionDataSource {
	res := &SessionDataSource{
		IntrospectionTableSource: func() *modconfig.ModResources {
			return w.GetResourceMaps()
		},
		PreparedStatementSource: func() *modconfig.ModResources {
			return w.GetResourceMaps()
		},
	}
	if preparedStatementSource != nil {
		res.PreparedStatementSource = func() *modconfig.ModResources {
			return preparedStatementSource
		}
	}
	return res
}
