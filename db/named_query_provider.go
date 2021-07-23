package db

import "github.com/turbot/steampipe/steampipeconfig/modconfig"

// NamedQueryProvider is an interface encapsulating named query searching capability
// - provided to avoid db needing a reference to workspace
type NamedQueryProvider interface {
	GetNamedQueryMap() map[string]*modconfig.Query
	GetNamedQuery(queryName string) (*modconfig.Query, bool)
}
