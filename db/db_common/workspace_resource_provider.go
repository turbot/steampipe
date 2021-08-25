package db_common

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// WorkspaceResourceProvider :: interface encapsulating named query searching capability
// - provided to avoid db needing a reference to workspace
type WorkspaceResourceProvider interface {
	GetQueryMap() map[string]*modconfig.Query
	GetQuery(queryName string) (*modconfig.Query, bool)
	GetControlMap() map[string]*modconfig.Control
	GetControl(controlName string) (*modconfig.Control, bool)
	SetupWatcher(client Client, onError func(err error)) error
}
