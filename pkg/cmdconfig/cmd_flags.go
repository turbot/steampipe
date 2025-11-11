package cmdconfig

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
)

var requiredColor = color.New(color.Bold).SprintfFunc()

type FlagOption func(c *cobra.Command, name string, key string)

// FlagOptions - shortcut for common flag options
var FlagOptions = struct {
	Required      func() FlagOption
	Hidden        func() FlagOption
	Deprecated    func(string) FlagOption
	NoOptDefVal   func(string) FlagOption
	WithShortHand func(string) FlagOption
}{
	Required:      requiredOpt,
	Hidden:        hiddenOpt,
	Deprecated:    deprecatedOpt,
	NoOptDefVal:   noOptDefValOpt,
	WithShortHand: withShortHand,
}

// Helper function to mark a flag as required
func requiredOpt() FlagOption {
	return func(c *cobra.Command, name, key string) {
		err := c.MarkFlagRequired(key)
		error_helpers.FailOnErrorWithMessage(err, "could not mark flag as required")
		key = fmt.Sprintf("required.%s", key)
		viperMutex.Lock()
		viper.GetViper().Set(key, true)
		viperMutex.Unlock()
		u := c.Flag(name).Usage
		c.Flag(name).Usage = fmt.Sprintf("%s %s", u, requiredColor("(required)"))
	}
}

func hiddenOpt() FlagOption {
	return func(c *cobra.Command, name, _ string) {
		c.Flag(name).Hidden = true
	}
}

func deprecatedOpt(replacement string) FlagOption {
	return func(c *cobra.Command, name, _ string) {
		c.Flag(name).Deprecated = fmt.Sprintf("please use %s", replacement)
	}
}

func noOptDefValOpt(noOptDefVal string) FlagOption {
	return func(c *cobra.Command, name, _ string) {
		c.Flag(name).NoOptDefVal = noOptDefVal
	}
}

func withShortHand(shorthand string) FlagOption {
	return func(c *cobra.Command, name, _ string) {
		c.Flag(name).Shorthand = shorthand
	}
}
