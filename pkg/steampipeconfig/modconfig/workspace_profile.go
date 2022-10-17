package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"reflect"
	"strings"
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

func (c *WorkspaceProfile) Initialise() hcl.Diagnostics {
	var diags hcl.Diagnostics
	var err error
	if c.InstallDir != "" {
		c.InstallDir, err = helpers.Tildefy(c.InstallDir)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{Severity: hcl.DiagError, Summary: err.Error()})
		}
	}

	if c.ModLocation != "" {
		c.ModLocation, err = helpers.Tildefy(c.ModLocation)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{Severity: hcl.DiagError, Summary: err.Error()})
		}
	}

	if c.snapshotLocationIsFilePath() {
		// so snapshot location _is_ file path
		// handle ~
		c.SnapshotLocation, err = files.Tildefy(c.SnapshotLocation)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{Severity: hcl.DiagError, Summary: err.Error()})
		}

		// ensure location exists
		if !files.DirectoryExists(c.SnapshotLocation) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("SnapshotLocation %s does not exist in local file system", c.SnapshotLocation),
			})
		}
	}
	return diags
}

// determine whether SnapshotLocation is a local path or a cloud workspace
// if it is a cloud workspace it will have the form {identity_handle}/{workspace_handle}
// otherwise we assume it is a local path
func (c *WorkspaceProfile) snapshotLocationIsFilePath() bool {
	if len(c.SnapshotLocation) == 0 {
		return false
	}
	parts := strings.Split(c.SnapshotLocation, "/")
	return len(parts) != 2
}
