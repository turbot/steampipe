package options

import (
	"fmt"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
)

type WorkspaceProfileDashboard struct {
	// workspace profile
	Browser *bool `hcl:"browser" cty:"profile_dashboard_browser"`
}

type GlobalDashboard struct {
	// server settings
	Port         *int    `hcl:"port"`
	Listen       *string `hcl:"listen"`
	StartTimeout *int    `hcl:"start_timeout"`
}

func (t *WorkspaceProfileDashboard) SetBaseProperties(otherOptions Options) {
	if helpers.IsNil(otherOptions) {
		return
	}
	if o, ok := otherOptions.(*WorkspaceProfileDashboard); ok {
		if t.Browser == nil && o.Browser != nil {
			t.Browser = o.Browser
		}
	}
}

// ConfigMap creates a config map that can be merged with viper
func (d *WorkspaceProfileDashboard) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if d.Browser != nil {
		res[constants.ArgBrowser] = d.Browser
	}
	return res
}

// Merge :: merge other options over the the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (d *WorkspaceProfileDashboard) Merge(otherOptions Options) {
	if _, ok := otherOptions.(*WorkspaceProfileDashboard); !ok {
		return
	}
	switch o := otherOptions.(type) {
	case *WorkspaceProfileDashboard:
		if o.Browser != nil {
			d.Browser = o.Browser
		}
	}
}

func (d *WorkspaceProfileDashboard) String() string {
	if d == nil {
		return ""
	}
	var str []string
	if d.Browser == nil {
		str = append(str, "  Browser: nil")
	} else {
		str = append(str, fmt.Sprintf("  Browser: %v", *d.Browser))
	}
	return strings.Join(str, "\n")
}

// ConfigMap creates a config map that can be merged with viper
func (d *GlobalDashboard) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if d.Port != nil {
		res[constants.ArgDashboardPort] = d.Port
	}
	if d.Listen != nil {
		res[constants.ArgDashboardListen] = d.Listen
	}
	if d.StartTimeout != nil {
		res[constants.ArgDashboardStartTimeout] = d.StartTimeout
	} else {
		res[constants.ArgDashboardStartTimeout] = constants.DashboardStartTimeout.Seconds()
	}
	return res
}

// Merge :: merge other options over the the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (d *GlobalDashboard) Merge(otherOptions Options) {
	if _, ok := otherOptions.(*GlobalDashboard); !ok {
		return
	}
	switch o := otherOptions.(type) {
	case *GlobalDashboard:
		if o.Port != nil {
			d.Port = o.Port
		}
		if o.Listen != nil {
			d.Listen = o.Listen
		}
		if o.StartTimeout != nil {
			d.StartTimeout = o.StartTimeout
		}
	}
}

func (d *GlobalDashboard) String() string {
	if d == nil {
		return ""
	}
	var str []string
	if d.Port == nil {
		str = append(str, "  Port: nil")
	} else {
		str = append(str, fmt.Sprintf("  Port: %d", *d.Port))
	}
	if d.Listen == nil {
		str = append(str, "  Listen: nil")
	} else {
		str = append(str, fmt.Sprintf("  Listen: %s", *d.Listen))
	}
	if d.StartTimeout == nil {
		str = append(str, "  StartTimeout: nil")
	} else {
		str = append(str, fmt.Sprintf("  StartTimeout: %d", *d.StartTimeout))
	}
	return strings.Join(str, "\n")
}
