package cmdconfig

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var requiredColor = color.New(color.Bold).SprintfFunc()

type flagOpt func(c *cobra.Command, name string, key string, v *viper.Viper)

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
	return func(c *cobra.Command, name, key string, v *viper.Viper) {
		c.MarkFlagRequired(key)
		key = fmt.Sprintf("required.%s", key)
		v.Set(key, true)
		u := c.Flag(name).Usage
		c.Flag(name).Usage = fmt.Sprintf("%s %s", u, requiredColor("(required)"))
	}
}

func hiddenOpt() flagOpt {
	return func(c *cobra.Command, name, key string, v *viper.Viper) {
		c.Flag(name).Hidden = true
	}
}
