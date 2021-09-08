package steampipeconfig

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type ConnectionMap map[string]*modconfig.Connection

func (c ConnectionMap) Equals(other ConnectionMap) bool {
	if len(c) != len(other) {
		return false
	}
	for k, v := range c {
		if !v.Equals(other[k]) {
			return false
		}
	}
	return true
}
