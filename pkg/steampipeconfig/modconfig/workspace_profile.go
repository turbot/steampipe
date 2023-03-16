package modconfig

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"github.com/zclconf/go-cty/cty"
)

type WorkspaceProfile struct {
	ProfileName       string            `hcl:"name,label" cty:"name"`
	CloudHost         *string           `hcl:"cloud_host,optional" cty:"cloud_host"`
	CloudToken        *string           `hcl:"cloud_token,optional" cty:"cloud_token"`
	InstallDir        *string           `hcl:"install_dir,optional" cty:"install_dir"`
	ModLocation       *string           `hcl:"mod_location,optional" cty:"mod_location"`
	SnapshotLocation  *string           `hcl:"snapshot_location,optional" cty:"snapshot_location"`
	WorkspaceDatabase *string           `hcl:"workspace_database,optional" cty:"workspace_database"`
	QueryTimeout      *int              `hcl:"query_timeout,optional" cty:"query_timeout"`
	Base              *WorkspaceProfile `hcl:"base"`
	SearchPath        *string           `hcl:"search_path" cty:"search_path"`
	SearchPathPrefix  *string           `hcl:"search_path_prefix" cty:"search_path_prefix"`
	Watch             *bool             `hcl:"watch" cty:"watch"`
	MaxParallel       *int              `hcl:"max_parallel" cty:"max-parallel"`

	// options
	QueryOptions      *options.Query
	CheckOptions      *options.Check
	GeneralOptions    *options.General
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
	case *options.General:
		if p.GeneralOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.GeneralOptions = o
	case *options.Query:
		if p.QueryOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.QueryOptions = o
	case *options.Check:
		if p.CheckOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.CheckOptions = o
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
	return GetCtyValue(p)
}

func (p *WorkspaceProfile) OnDecoded() hcl.Diagnostics {
	p.setBaseProperties()
	return nil
}

func (p *WorkspaceProfile) setBaseProperties() {
	if p.Base == nil {
		return
	}

	if p.CloudHost == nil {
		p.CloudHost = p.Base.CloudHost
	}
	if p.CloudToken == nil {
		p.CloudToken = p.Base.CloudToken
	}
	if p.InstallDir == nil {
		p.InstallDir = p.Base.InstallDir
	}
	if p.ModLocation == nil {
		p.ModLocation = p.Base.ModLocation
	}
	if p.SnapshotLocation == nil {
		p.SnapshotLocation = p.Base.SnapshotLocation
	}
	if p.WorkspaceDatabase == nil {
		p.WorkspaceDatabase = p.Base.WorkspaceDatabase
	}
	if p.QueryTimeout == nil {
		p.QueryTimeout = p.Base.QueryTimeout
	}
	if p.SearchPath == nil {
		p.SearchPath = p.Base.SearchPath
	}
	if p.SearchPathPrefix == nil {
		p.SearchPathPrefix = p.Base.SearchPathPrefix
	}
	if p.Watch == nil {
		p.Watch = p.Base.Watch
	}
	if p.MaxParallel == nil {
		p.MaxParallel = p.Base.MaxParallel
	}
}

// ConfigMap creates a config map containing all options to pass to viper
func (p *WorkspaceProfile) ConfigMap(commandName string) map[string]interface{} {
	res := ConfigMap{}
	// add non-empty properties to config map

	res.SetStringItem(p.CloudHost, constants.ArgCloudHost)
	res.SetStringItem(p.CloudToken, constants.ArgCloudToken)
	res.SetStringItem(p.InstallDir, constants.ArgInstallDir)
	res.SetStringItem(p.ModLocation, constants.ArgModLocation)
	res.SetStringItem(p.SnapshotLocation, constants.ArgSnapshotLocation)
	res.SetStringItem(p.WorkspaceDatabase, constants.ArgWorkspaceDatabase)
	res.SetIntItem(p.QueryTimeout, constants.ArgDatabaseQueryTimeout)
	res.SetBoolItem(p.Watch, constants.ArgWatch)
	res.SetIntItem(p.MaxParallel, constants.ArgMaxParallel)
	res.SetStringSliceItem(searchPathToArray(*p.SearchPath), constants.ArgSearchPath)
	res.SetStringSliceItem(searchPathToArray(*p.SearchPathPrefix), constants.ArgSearchPathPrefix)

	// now add options
	// build flat config map with order or precedence (low to high): general, terminal, connection
	// this means if (for example) 'search-path' is set in both terminal and connection options,
	// the value from connection options will have precedence
	// however, we also store all values scoped by their options type, so we will store:
	// 'database.search-path', 'terminal.search-path' AND 'search-path' (which will be equal to 'terminal.search-path')
	if p.GeneralOptions != nil {
		res.PopulateConfigMapForOptions(p.GeneralOptions)
	}
	if p.ConnectionOptions != nil {
		res.PopulateConfigMapForOptions(p.ConnectionOptions)
	}
	if commandName == "query" && p.QueryOptions != nil {
		res.PopulateConfigMapForOptions(p.QueryOptions)
	}
	if commandName == "check" && p.CheckOptions != nil {
		res.PopulateConfigMapForOptions(p.CheckOptions)
	}

	return res
}

func searchPathToArray(searchPathString string) []string {
	// convert comma separated list to array
	searchPath := strings.Split(searchPathString, ",")
	// strip whitespace
	for i, s := range searchPath {
		searchPath[i] = strings.TrimSpace(s)
	}
	return searchPath
}
