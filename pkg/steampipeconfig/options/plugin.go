package options

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
)

type Plugin struct {
	MemoryMaxMb *int `hcl:"memory_max_mb"`
}

// ConfigMap creates a config map that can be merged with viper
func (t *Plugin) ConfigMap() map[string]interface{} {
	// only add keys which are non-null
	res := map[string]interface{}{}
	if t.MemoryMaxMb != nil {
		res[constants.ArgMemoryMaxMbPlugin] = t.MemoryMaxMb
	}

	return res
}

// Merge merges other options over the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (t *Plugin) Merge(otherOptions Options) {
	switch o := otherOptions.(type) {
	case *Plugin:
		if o.MemoryMaxMb != nil {
			t.MemoryMaxMb = o.MemoryMaxMb
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

	return strings.Join(str, "\n")
}
