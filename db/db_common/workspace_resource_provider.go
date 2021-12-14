package db_common

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// WorkspaceResourceProvider is an interface encapsulating workspace functionality
// - provided to avoid db needing a reference to Workspace
type WorkspaceResourceProvider interface {
	ResolveQueryAndArgs(arg string) (string, modconfig.QueryProvider, error)
	GetQueryMap() map[string]*modconfig.Query
	GetControlMap() map[string]*modconfig.Control
	GetResourceMaps() *modconfig.WorkspaceResourceMaps
	GetControl(controlName string) (*modconfig.Control, bool)
	SetupWatcher(client Client, onError func(err error)) error
	SetOnFileWatcherEventMessages(f func())
}
