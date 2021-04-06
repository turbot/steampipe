package options

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/constants"
)

// General
type General struct {
	LogLevel    *string `hcl:"log_level"`
	UpdateCheck *string `hcl:"update_check"`
}

// ConfigMap :: create a config map to pass to viper
func (g *General) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if g.LogLevel != nil {
		res[constants.ArgLogLevel] = g.LogLevel
	}
	if g.UpdateCheck != nil {
		res[constants.ArgUpdateCheck] = g.UpdateCheck
	}

	return res
}

// merge other options over the the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (g *General) Merge(otherOptions Options) {
	switch o := otherOptions.(type) {
	case *General:
		if o.LogLevel != nil {
			g.LogLevel = o.LogLevel
		}
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
	if g.LogLevel == nil {
		str = append(str, "LogLevel: nil")
	} else {
		str = append(str, fmt.Sprintf("LogLevel: %s", *g.LogLevel))
	}
	if g.UpdateCheck == nil {
		str = append(str, "UpdateCheck: nil")
	} else {
		str = append(str, fmt.Sprintf("UpdateCheck: %s", *g.UpdateCheck))
	}
	return strings.Join(str, "\n")
}
