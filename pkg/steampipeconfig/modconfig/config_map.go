package modconfig

import (
	"fmt"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"reflect"
	"strings"
)

type ConfigMap map[string]interface{}

// SetStringItem checks is string is non-empty and if so, add to map with given key
func (m ConfigMap) SetStringItem(argValue string, argName string) {
	if argValue != "" {
		m[argName] = argValue
	}
}

// PopulateConfigMapForOptions populates the config map for a given options object
// NOTE: this mutates configMap
func (m ConfigMap) PopulateConfigMapForOptions(o options.Options) {
	for k, v := range o.ConfigMap() {
		// skip empty string
		if s, ok := v.(string); ok && s == "" {
			continue
		}
		m[k] = v
		// also store a scoped version of the config property
		m[getScopedKey(o, k)] = v
	}
}

// generated a scoped key for the config property. For example if o is a database options object and k is 'search-path'
// the scoped key will be 'database.search-path'
func getScopedKey(o options.Options, k string) string {
	t := reflect.TypeOf(helpers.DereferencePointer(o)).Name()
	return fmt.Sprintf("%s.%s", strings.ToLower(t), k)
}
