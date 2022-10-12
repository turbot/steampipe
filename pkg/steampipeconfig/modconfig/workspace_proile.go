package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"reflect"
)

type WorkspaceProfile struct {
	Name              string `hcl:"name,label"`
	CloudToken        string `hcl:"cloud_token,optional"`
	CloudHost         string `hcl:"cloud_host,optional"`
	WorkspaceDatabase string `hcl:"workspace_database,optional"`
	SnapshotLocation  string `hcl:"snapshot_location,optional"`
	ModLocation       string `hcl:"mod_location,optional"`
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

//func (c *WorkspaceProfile) String() string {
//	return fmt.Sprintf("\n----\nName: %s\nPlugin: %s\nConfig:\n%s\nOptions:\n%s\n", c.Name, c.Plugin, c.Config, c.Options.String())
//}

//// Validate verifies the Type property is valid,
//// if this is an aggregator connection, there must be at least one child, and no duplicates
//// if this is NOT an aggregator, there must be no children
//func (c *WorkspaceProfile) Validate(connectionMap map[string]*WorkspaceProfile) []string {
//	validConnectionTypes := []string{"", ConnectionTypeAggregator}
//	if !helpers.StringSliceContains(validConnectionTypes, c.Type) {
//		return []string{fmt.Sprintf("connection '%s' has invalid connection type '%s'", c.Name, c.Type)}
//	}
//	if c.Type == ConnectionTypeAggregator {
//		return c.ValidateAggregatorConnection(connectionMap)
//	}
//	// this is NOT an aggregator group - there should be no children
//	var validationErrors []string
//
//	if len(c.ConnectionNames) != 0 {
//		validationErrors = append(validationErrors, fmt.Sprintf("connection '%s' has %d children, but is not of type 'aggregator'", c.Name, len(c.ConnectionNames)))
//	}
//	return validationErrors
//
//}
