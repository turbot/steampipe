package plugin

import "github.com/turbot/pipe-fittings/v2/hclhelpers"

type PluginConnection interface {
	GetDeclRange() hclhelpers.Range
	GetName() string
	GetDisplayName() string
}
