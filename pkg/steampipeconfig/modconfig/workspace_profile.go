package modconfig

import "C"
import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"github.com/zclconf/go-cty/cty"
	"reflect"
)

type WorkspaceProfile struct {
	ProfileName       string            `hcl:"name,label" cty:"name"`
	CloudHost         string            `hcl:"cloud_host,optional" cty:"cloud_host"`
	CloudToken        string            `hcl:"cloud_token,optional" cty:"cloud_token"`
	InstallDir        string            `hcl:"install_dir,optional" cty:"install_dir"`
	ModLocation       string            `hcl:"mod_location,optional" cty:"mod_location"`
	SnapshotLocation  string            `hcl:"snapshot_location,optional" cty:"snapshot_location"`
	WorkspaceDatabase string            `hcl:"workspace_database,optional" cty:"workspace_database"`
	Base              *WorkspaceProfile `hcl:"base"`

	// options
	ConnectionOptions *options.Connection
	TerminalOptions   *options.Terminal
	GeneralOptions    *options.General
	DeclRange         hcl.Range
}

func NewWorkspaceProfile(block *hcl.Block) *WorkspaceProfile {
	return &WorkspaceProfile{
		ProfileName: block.Labels[0],
		DeclRange:   block.TypeRange,
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

func (c *WorkspaceProfile) Name() string {
	return c.Name()
}
func (c *WorkspaceProfile) GetUnqualifiedName() string {
	return c.Name()
}
func (c *WorkspaceProfile) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

func (c *WorkspaceProfile) OnDecoded(block *hcl.Block, resourceMapProvider ResourceMapsProvider) hcl.Diagnostics {
	c.setBaseProperties(resourceMapProvider)
	return nil
}

func (c *WorkspaceProfile) AddReference(*ResourceReference)     {}
func (c *WorkspaceProfile) GetReferences() []*ResourceReference { return nil }
func (c *WorkspaceProfile) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

func (d *WorkspaceProfile) setBaseProperties(resourceMapProvider ResourceMapsProvider) {
	//// not all base properties are stored in the evalContext
	//// (e.g. resource metadata and runtime dependencies are not stores)
	////  so resolve base from the resource map provider (which is the RunContext)
	//if base, resolved := resolveBase(d.Base, resourceMapProvider); !resolved {
	//	return
	//} else {
	//	d.Base = base.(*Dashboard)
	//}
	//
	//if d.Title == nil {
	//	d.Title = d.Base.Title
	//}
	//
	//if d.Width == nil {
	//	d.Width = d.Base.Width
	//}
	//
	//if len(d.children) == 0 {
	//	d.children = d.Base.children
	//	d.ChildNames = d.Base.ChildNames
	//}
	//
	//d.addBaseInputs(d.Base.Inputs)
	//
	//d.Tags = utils.MergeMaps(d.Tags, d.Base.Tags)
	//
	//if d.Description == nil {
	//	d.Description = d.Base.Description
	//}
	//
	//if d.Documentation == nil {
	//	d.Documentation = d.Base.Documentation
	//}
}
