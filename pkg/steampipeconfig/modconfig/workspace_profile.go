package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/constants"
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
	GeneralOptions    *options.General
	TerminalOptions   *options.Terminal
	ConnectionOptions *options.Connection
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
		if p.ConnectionOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.ConnectionOptions = o
	case *options.Terminal:
		if p.TerminalOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.TerminalOptions = o
	case *options.General:
		if p.GeneralOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
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

func duplicateOptionsBlockDiag(block *hcl.Block) *hcl.Diagnostic {
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("duplicate %s options block", block.Type),
		Subject:  &block.DefRange,
	}
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

// ConfigMap creates a config map containing all options to pass to viper
func (p *WorkspaceProfile) ConfigMap() map[string]interface{} {
	res := ConfigMap{}
	// add non-empty properties to config map

	res.SetStringItem(p.CloudHost, constants.ArgCloudHost)
	res.SetStringItem(p.CloudToken, constants.ArgCloudToken)
	res.SetStringItem(p.InstallDir, constants.ArgInstallDir)
	res.SetStringItem(p.ModLocation, constants.ArgModLocation)
	res.SetStringItem(p.SnapshotLocation, constants.ArgSnapshotLocation)
	res.SetStringItem(p.WorkspaceDatabase, constants.ArgWorkspaceDatabase)

	// now add options
	// build flat config map with order or precedence (low to high): general, terminal, connection
	// this means if (for example) 'search-path' is set in both terminal and connection options,
	// the value from connection options will have precedence
	// however, we also store all values scoped by their options type, so we will store:
	// 'database.search-path', 'terminal.search-path' AND 'search-path' (which will be equal to 'terminal.search-path')
	if p.GeneralOptions != nil {
		res.PopulateConfigMapForOptions(p.GeneralOptions)
	}
	if p.TerminalOptions != nil {
		res.PopulateConfigMapForOptions(p.TerminalOptions)
	}
	if p.ConnectionOptions != nil {
		res.PopulateConfigMapForOptions(p.ConnectionOptions)
	}

	return res
}
