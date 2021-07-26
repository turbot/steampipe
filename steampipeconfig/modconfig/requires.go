package modconfig

import "github.com/hashicorp/hcl/v2"

// Requires is a struct representing mod dependencies
type Requires struct {
	Steampipe string           `hcl:"steampipe,optional"`
	Plugins   []*PluginVersion `hcl:"plugin,block"`
	Mods      []*ModVersion    `hcl:"mod,block"`
	DeclRange hcl.Range        `json:"-"`
}
