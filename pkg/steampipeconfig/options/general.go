package options

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
)

type General struct {
	UpdateCheck *string `hcl:"update_check"`
	MaxParallel *int    `hcl:"max_parallel"`
	Telemetry   *string `hcl:"telemetry"`
	LogLevel    *string `hcl:"log_level"`
	MemoryMaxMb *int    `hcl:"memory_max_mb"`
}

// ConfigMap creates a config map that can be merged with viper
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
	if g.LogLevel != nil {
		res[constants.ArgLogLevel] = g.LogLevel
	}
	if g.MemoryMaxMb != nil {
		res[constants.ArgMemoryMaxMb] = g.MemoryMaxMb
	}

	return res
}

// Merge merges other options over the top of this options object
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

	if g.MaxParallel == nil {
		str = append(str, "  MaxParallel: nil")
	} else {
		str = append(str, fmt.Sprintf("  MaxParallel: %d", *g.MaxParallel))
	}

	if g.Telemetry == nil {
		str = append(str, "  Telemetry: nil")
	} else {
		str = append(str, fmt.Sprintf("  Telemetry: %s", *g.Telemetry))
	}
	if g.LogLevel == nil {
		str = append(str, "  LogLevel: nil")
	} else {
		str = append(str, fmt.Sprintf("  LogLevel: %s", *g.LogLevel))
	}

	if g.MemoryMaxMb == nil {
		str = append(str, "  MemoryMaxMb: nil")
	} else {
		str = append(str, fmt.Sprintf("  MemoryMaxMb: %d", *g.MemoryMaxMb))
	}
	return strings.Join(str, "\n")
}
