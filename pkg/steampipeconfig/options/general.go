package options

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
)

// General
type General struct {
	UpdateCheck *string `hcl:"update_check"`
	MaxParallel *int    `hcl:"max_parallel"`
	Telemetry   *string `hcl:"telemetry"`
}

// ConfigMap :: create a config map to pass to viper
func (g *General) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if g.UpdateCheck != nil {
		res[constants.ArgUpdateCheck] = g.UpdateCheck
	}
	if g.Telemetry != nil {
		res[constants.ArgTelemetry] = g.Telemetry
	}
	if g.MaxParallel != nil {
		res[constants.ArgMaxParallel] = g.MaxParallel
	}

	return res
}

// Merge :: merge other options over the the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (g *General) Merge(otherOptions Options) {
	switch o := otherOptions.(type) {
	case *General:
		if o.UpdateCheck != nil {
			g.UpdateCheck = o.UpdateCheck
		}
	}
}

func (g *General) String() string {
	if g == nil {
		return ""
	}
	var str []string
	if g.UpdateCheck == nil {
		str = append(str, "  UpdateCheck: nil")
	} else {
		str = append(str, fmt.Sprintf("  UpdateCheck: %s", *g.UpdateCheck))
	}
	return strings.Join(str, "\n")
}
