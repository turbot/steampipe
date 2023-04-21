package modconfig

import (
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
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
	}
}
