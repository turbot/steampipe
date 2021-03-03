package cmdconfig

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var requiredColor = color.New(color.Bold).SprintfFunc()

type flagOpt func(c *cobra.Command, name string, key string, v *ViperWrapper)

// FlagOptions :: shortcut for common flag options
var FlagOptions = struct {
	Required func() flagOpt
	Hidden   func() flagOpt
}{
	Required: requiredOpt,
	Hidden:   hiddenOpt,
}

// Helper function to mark a flag as required
func requiredOpt() flagOpt {
	return func(c *cobra.Command, name, key string, w *ViperWrapper) {
		c.MarkFlagRequired(key)
		key = fmt.Sprintf("required.%s", key)
		w.v.Set(key, true)
		u := c.Flag(name).Usage
		c.Flag(name).Usage = fmt.Sprintf("%s %s", u, requiredColor("(required)"))
	}
}

func hiddenOpt() flagOpt {
	return func(c *cobra.Command, name, key string, v *ViperWrapper) {
		c.Flag(name).Hidden = true
	}
}
