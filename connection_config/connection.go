package connection_config

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
)

// Connection :: structure representing the partially parsed connection.
type Connection struct {
	// connection name
	Name string
	// FQN of plugin
	Plugin string
	// unparsed HCL of plugin specific connection config
	Config string

	// options
	Options *ConnectionOptions
}

// set the options on the connection
// verify the options is a valid options type (only ConnectionOptions currently supported)
func (c *Connection) setOptions(options Options, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch o := options.(type) {
	case *ConnectionOptions:
		c.Options = o
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Connections", reflect.TypeOf(options).Name()),
			Subject:  &block.DefRange,
		})
	}
	return diags
}
