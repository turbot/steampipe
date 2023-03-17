package modconfig

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
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
	QueryTimeout      *int              `hcl:"query_timeout,optional" cty:"query_timeout"`
	SnapshotLocation  *string           `hcl:"snapshot_location,optional" cty:"snapshot_location"`
	WorkspaceDatabase *string           `hcl:"workspace_database,optional" cty:"workspace_database"`
	SearchPath        *string           `hcl:"search_path" cty:"search_path"`
	SearchPathPrefix  *string           `hcl:"search_path_prefix" cty:"search_path_prefix"`
	Watch             *bool             `hcl:"watch" cty:"watch"`
	MaxParallel       *int              `hcl:"max_parallel" cty:"max-parallel"`
	Introspection     *bool             `hcl:"introspection" cty:"introspection"`
	Input             *bool             `hcl:"input" cty:"input"`
	Progress          *bool             `hcl:"progress" cty:"progress"`
	Theme             *string           `hcl:"theme" cty:"theme"`
	Cache             *bool             `hcl:"cache" cty:"cache"`
	CacheTTL          *int              `hcl:"cache_ttl" cty:"cache_ttl"`
	Base              *WorkspaceProfile `hcl:"base"`
	SearchPath        *string           `hcl:"search_path" cty:"search_path"`
	SearchPathPrefix  *string           `hcl:"search_path_prefix" cty:"search_path_prefix"`
	Watch             *bool             `hcl:"watch" cty:"watch"`
	MaxParallel       *int              `hcl:"max_parallel" cty:"max-parallel"`
	Introspection     *bool             `hcl:"introspection" cty:"introspection"`
	Input             *bool             `hcl:"input" cty:"input"`
	Progress          *bool             `hcl:"progress" cty:"progress"`
	Theme             *string           `hcl:"theme" cty:"theme"`
	Cache             *bool             `hcl:"cache" cty:"cache"`
	CacheTTL          *int              `hcl:"cache_ttl" cty:"cache_ttl"`
	Base              *WorkspaceProfile `hcl:"base"`

	// options
	QueryOptions     *options.Query                     `cty:"query-options"`
	CheckOptions     *options.Check                     `cty:"check-options"`
	DashboardOptions *options.WorkspaceProfileDashboard `cty:"dashboard-options"`
	DeclRange        hcl.Range
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
	case *options.WorkspaceProfileDashboard:
		if p.DashboardOptions != nil {
			diags = append(diags, duplicateOptionsBlockDiag(block))
		}
		p.DashboardOptions = o
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
	if p.Introspection == nil {
		p.Introspection = p.Base.Introspection
	}
	if p.Input == nil {
		p.Input = p.Base.Input
	}
	if p.Progress == nil {
		p.Progress = p.Base.Progress
	}
	if p.Theme == nil {
		p.Theme = p.Base.Theme
	}
	if p.Cache == nil {
		p.Cache = p.Base.Cache
	}
	if p.CacheTTL == nil {
		p.CacheTTL = p.Base.CacheTTL
	}

	// nested inheritance strategy:
	//
	// if my nested struct is a nil
	//		-> use the base struct
	//
	// if I am not nil (and base is not nil)
	//		-> only inherit the properties which are nil in me and not in base
	//
	if p.QueryOptions == nil {
		p.QueryOptions = p.Base.QueryOptions
	} else {
		p.QueryOptions.SetBaseProperties(p.Base.QueryOptions)
	}
	if p.CheckOptions == nil {
		p.CheckOptions = p.Base.CheckOptions
	} else {
		p.CheckOptions.SetBaseProperties(p.Base.CheckOptions)
	}
	if p.DashboardOptions == nil {
		p.DashboardOptions = p.Base.DashboardOptions
	} else {
		p.DashboardOptions.SetBaseProperties(p.Base.DashboardOptions)
	}
}

// ConfigMap creates a config map containing all options to pass to viper
func (p *WorkspaceProfile) ConfigMap(cmd *cobra.Command) map[string]interface{} {
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
	res.SetStringSliceItem(searchPathFromString(p.SearchPath, ","), constants.ArgSearchPath)
	res.SetStringSliceItem(searchPathFromString(p.SearchPathPrefix, ","), constants.ArgSearchPathPrefix)
	res.SetBoolItem(p.Introspection, constants.ArgIntrospection)
	res.SetBoolItem(p.Input, constants.ArgInput)
	res.SetBoolItem(p.Progress, constants.ArgProgress)
	res.SetStringItem(p.Theme, constants.ArgTheme)
	res.SetBoolItem(p.Cache, constants.ArgCache)
	res.SetIntItem(p.CacheTTL, constants.ArgCacheTtl)

	if cmd.Name() == constants.CmdNameQuery && p.QueryOptions != nil {
		res.PopulateConfigMapForOptions(p.QueryOptions)
	}
	if cmd.Name() == constants.CmdNameCheck && p.CheckOptions != nil {
		res.PopulateConfigMapForOptions(p.CheckOptions)
	}
	if cmd.Name() == constants.CmdNameDashboard && p.DashboardOptions != nil {
		res.PopulateConfigMapForOptions(p.DashboardOptions)
	}
	if commandName == "query" && p.QueryOptions != nil {
		res.PopulateConfigMapForOptions(p.QueryOptions)
	}
	if cmd.Name() == "check" && p.CheckOptions != nil {
		res.PopulateConfigMapForOptions(p.CheckOptions)
	}
	if cmd.Name() == "dashboard" && p.DashboardOptions != nil {
		res.PopulateConfigMapForOptions(p.DashboardOptions)
	}

	return res
}

// searchPathFromString checks that `str` is `nil` and returns a string slice with `str`
// separated with `separator`
// If `str` is `nil`, this returns a `nil`
func searchPathFromString(str *string, separator string) []string {
	if str == nil {
		return nil
	}
	// convert comma separated list to array
	searchPath := strings.Split(*str, separator)
	// strip whitespace
	for i, s := range searchPath {
		searchPath[i] = strings.TrimSpace(s)
	}
	return searchPath
}
