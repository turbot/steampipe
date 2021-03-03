package cmdconfig

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
)

var viperWrapper *ViperWrapper

// InitViper :: initializes and configures an instance of viper
func InitViper(w *ViperWrapper) {
	w.v.SetEnvPrefix("STEAMPIPE")
	w.v.AutomaticEnv()
	w.v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// set defaults
	w.Set(constants.ShowInteractiveOutputConfigKey, true)
}

// sets a global viper instance
func setConfig(v *ViperWrapper) {
	viperWrapper = v
}

// Viper :: fetches the global viper instance
func Viper() *ViperWrapper {
	return viperWrapper
}

type ViperWrapper struct {
	v      *viper.Viper
	prefix string
}

func NewViperWrapper(cmd *cobra.Command) *ViperWrapper {
	w := new(ViperWrapper)
	w.v = viper.GetViper()
	w.prefix = fmt.Sprintf("%s_%p", cmd.Use, cmd)
	return w
}
func (w *ViperWrapper) GetString(key string) string {
	return w.v.GetString(w.buildGetKey(key))
}
func (w *ViperWrapper) GetBool(key string) bool {
	return w.v.GetBool(w.buildGetKey(key))
}
func (w *ViperWrapper) GetInt(key string) int {
	return w.v.GetInt(w.buildGetKey(key))
}
func (w *ViperWrapper) Set(key string, value interface{}) {
	w.v.Set(w.buildSetKey(key), value)
}
func (w *ViperWrapper) BindPFlag(key string, flag *pflag.Flag) error {
	newKey := w.buildGetKey(key)
	return w.v.BindPFlag(newKey, flag)
}
func (w *ViperWrapper) buildSetKey(key string) string {
	return fmt.Sprintf("%s_%s", w.prefix, key)
}
func (w *ViperWrapper) buildGetKey(key string) string {
	prefixedKey := fmt.Sprintf("%s_%s", w.prefix, key)
	if w.v.IsSet(prefixedKey) {
		return prefixedKey
	}
	return key
}
