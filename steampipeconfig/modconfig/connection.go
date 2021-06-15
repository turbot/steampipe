package modconfig

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/steampipeconfig/options"
)

// Connection is a struct representing the partially parsed connection
//
// (Partial as the connection config, which is plugin specific, is stored as raw HCL.
// This will be parsed by the plugin)
type Connection struct {
	// connection name
	Name string
	// Name of plugin
	Plugin string
	// unparsed HCL of plugin specific connection config
	Config string

	// options
	Options *options.Connection
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (c *Connection) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch o := opts.(type) {
	case *options.Connection:
		c.Options = o
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Connections", reflect.TypeOf(o).Name()),
			Subject:  &block.DefRange,
		})
	}
	return diags
}

func (c *Connection) String() string {
	return fmt.Sprintf("\n----\nName: %s\nPlugin: %s\nConfig:\n%s\nOptions:\n%s\n", c.Name, c.Plugin, c.Config, c.Options.String())
}
