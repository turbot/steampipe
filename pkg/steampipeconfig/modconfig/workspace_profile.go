package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"reflect"
)

type WorkspaceProfile struct {
	Name              string `hcl:"name,label"`
	CloudHost         string `hcl:"cloud_host,optional"`
	CloudToken        string `hcl:"cloud_token,optional"`
	InstallDir        string `hcl:"install_dir,optional"`
	ModLocation       string `hcl:"mod_location,optional"`
	SnapshotLocation  string `hcl:"snapshot_location,optional"`
	WorkspaceDatabase string `hcl:"workspace_database,optional"`
	//Base      	 *WorkspaceProfile `hcl:"base"`

	// options
	ConnectionOptions *options.Connection
	TerminalOptions   *options.Terminal
	GeneralOptions    *options.General
	DeclRange         hcl.Range
}

func NewWorkspaceProfile(block *hcl.Block) *WorkspaceProfile {
	return &WorkspaceProfile{
		Name:      block.Labels[0],
		DeclRange: block.TypeRange,
	}
}

// SetOptions sets the options on the connection
// verify the options object is a valid options type (only options.Connection currently supported)
func (c *WorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch o := opts.(type) {
	case *options.Connection:
		c.ConnectionOptions = o
	case *options.Terminal:
		c.TerminalOptions = o
	case *options.General:
		c.GeneralOptions = o
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Connections", reflect.TypeOf(o).Name()),
			Subject:  &block.DefRange,
		})
	}
	return diags
}
