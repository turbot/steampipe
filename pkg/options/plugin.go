package options

import (
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/options"
)

type Plugin struct {
	MemoryMaxMb  *int `hcl:"memory_max_mb"`
	StartTimeout *int `hcl:"start_timeout"`
}

// ConfigMap creates a config map that can be merged with viper
func (t *Plugin) ConfigMap() map[string]interface{} {
	// only add keys which are non-null
	res := map[string]interface{}{}
	if t.MemoryMaxMb != nil {
		res[constants.ArgMemoryMaxMbPlugin] = t.MemoryMaxMb
	}
	if t.StartTimeout != nil {
		res[constants.ArgPluginStartTimeout] = t.StartTimeout
	}

	return res
}

// Merge merges other options over the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (t *Plugin) Merge(otherOptions options.Options) {
	switch o := otherOptions.(type) {
	case *Plugin:
		if o.MemoryMaxMb != nil {
			t.MemoryMaxMb = o.MemoryMaxMb
		}
		if o.StartTimeout != nil {
			t.StartTimeout = o.StartTimeout
		}
	}
}

func (t *Plugin) String() string {
	if t == nil {
		return ""
	}
	var str []string
	if t.MemoryMaxMb == nil {
		str = append(str, "  MemoryMaxMb: nil")
	} else {
		str = append(str, fmt.Sprintf("  MemoryMaxMb: %d", *t.MemoryMaxMb))
	}
	if t.StartTimeout == nil {
		str = append(str, "  PluginStartTimeout: nil")
	} else {
		str = append(str, fmt.Sprintf("  PluginStartTimeout: %d", *t.StartTimeout))
	}

	return strings.Join(str, "\n")
}
