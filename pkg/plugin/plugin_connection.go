package plugin

import "github.com/turbot/pipe-fittings/hclhelpers"

type PluginConnection interface {
	GetDeclRange() hclhelpers.Range
	GetName() string
	GetDisplayName() string
}
