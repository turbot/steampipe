package db_common

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// WorkspaceResourceProvider :: interface encapsulating named query searching capability
// - provided to avoid db needing a reference to workspace
type WorkspaceResourceProvider interface {
	ResolveQueryAndArgs(arg string) (string, error)
	GetQueryMap() map[string]*modconfig.Query
	GetControlMap() map[string]*modconfig.Control
	GetResourceMaps() *modconfig.WorkspaceResourceMaps
	GetControl(controlName string) (*modconfig.Control, bool)
	SetupWatcher(client Client, onError func(err error)) error
}
