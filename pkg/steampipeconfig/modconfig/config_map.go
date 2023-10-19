package modconfig

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/options"
)

type ConfigMap map[string]interface{}

// SetStringItem checks is string pointer is non-nil and if so, add to map with given key
func (m ConfigMap) SetStringItem(argValue *string, argName string) {
	if argValue != nil {
		m[argName] = *argValue
	}
}

// SetStringSliceItem checks is string slice pointer is non-nil and if so, add to map with given key
func (m ConfigMap) SetStringSliceItem(argValue []string, argName string) {
	if argValue != nil {
		m[argName] = argValue
	}
}

// SetIntItem checks is int pointer is non-nil and if so, add to map with given key
func (m ConfigMap) SetIntItem(argValue *int, argName string) {
	if argValue != nil {
		m[argName] = *argValue
	}
}

// SetBoolItem checks is bool pointer is non-nil and if so, add to map with given key
func (m ConfigMap) SetBoolItem(argValue *bool, argName string) {
	if argValue != nil {
		m[argName] = *argValue
	}
}

// PopulateConfigMapForOptions populates the config map for a given options object
// NOTE: this mutates configMap
func (m ConfigMap) PopulateConfigMapForOptions(o options.Options) {
	for k, v := range o.ConfigMap() {
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
