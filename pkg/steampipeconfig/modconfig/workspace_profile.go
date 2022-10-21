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
func (p *WorkspaceProfile) SetOptions(opts options.Options, block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch o := opts.(type) {
	case *options.Connection:
		p.ConnectionOptions = o
	case *options.Terminal:
		p.TerminalOptions = o
	case *options.General:
		p.GeneralOptions = o
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid nested option type %s - only 'connection' options blocks are supported for Connections", reflect.TypeOf(o).Name()),
			Subject:  &block.DefRange,
		})
	}
	return diags
}

func (p *WorkspaceProfile) Name() string {
	return fmt.Sprintf("workspace.%s", p.ProfileName)
}

func (p *WorkspaceProfile) CtyValue() (cty.Value, error) {
	return getCtyValue(p)
}

func (p *WorkspaceProfile) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

//func (c *WorkspaceProfile) AddReference(*ResourceReference)     {}
//func (c *WorkspaceProfile) GetReferences() []*ResourceReference { return nil }
//func (c *WorkspaceProfile) GetDeclRange() *hcl.Range {
//	return &c.DeclRange
//}

func (p *WorkspaceProfile) setBaseProperties() {
	if p.Base == nil {
		return
	}

	if p.CloudHost == "" {
		p.CloudHost = p.Base.CloudHost
	}
	if p.CloudToken == "" {
		p.CloudToken = p.Base.CloudToken
	}
	if p.InstallDir == "" {
		p.InstallDir = p.Base.InstallDir
	}
	if p.ModLocation == "" {
		p.ModLocation = p.Base.ModLocation
	}
	if p.SnapshotLocation == "" {
		p.SnapshotLocation = p.Base.SnapshotLocation
	}
	if p.WorkspaceDatabase == "" {
		p.WorkspaceDatabase = p.Base.WorkspaceDatabase
	}
}
