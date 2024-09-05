package plugin

import pplugin "github.com/turbot/pipe-fittings/plugin"

type PluginConnection interface {
	GetDeclRange() pplugin.Range
	GetName() string
	GetDisplayName() string
}
