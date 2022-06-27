package cmdconfig

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var requiredColor = color.New(color.Bold).SprintfFunc()

type flagOpt func(c *cobra.Command, name string, key string)

// FlagOptions :: shortcut for common flag options
var FlagOptions = struct {
	Required   func() flagOpt
	Hidden     func() flagOpt
	Deprecated func(string) flagOpt
}{
	Required:   requiredOpt,
	Hidden:     hiddenOpt,
	Deprecated: deprecatedOpt,
}

// Helper function to mark a flag as required
func requiredOpt() flagOpt {
	return func(c *cobra.Command, name, key string) {
		c.MarkFlagRequired(key)
		key = fmt.Sprintf("required.%s", key)
		viper.GetViper().Set(key, true)
		u := c.Flag(name).Usage
		c.Flag(name).Usage = fmt.Sprintf("%s %s", u, requiredColor("(required)"))
	}
}

func hiddenOpt() flagOpt {
	return func(c *cobra.Command, name, key string) {
		c.Flag(name).Hidden = true
	}
}

func deprecatedOpt(replacement string) flagOpt {
	return func(c *cobra.Command, name, key string) {
		c.Flag(name).Deprecated = fmt.Sprintf("please use %s", replacement)
	}
}
