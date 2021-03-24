package options

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/constants"
)

// Database
type Database struct {
	Port   *int    `hcl:"port"`
	Listen *string `hcl:"listen"`
}

// ConfigMap :: create a config map to pass to viper
func (c *Database) ConfigMap() map[string]interface{} {
	// only add keys which are non null
	res := map[string]interface{}{}
	if c.Port != nil {
		res[constants.ArgPort] = c.Port
	}
	if c.Listen != nil {
		res[constants.ArgListenAddress] = c.Listen
	}
	return res
}

func (c *Database) String() string {
	if c == nil {
		return ""
	}
	var str []string
	if c.Port == nil {
		str = append(str, "Port: nil")
	} else {
		str = append(str, fmt.Sprintf("Port: %d", *c.Port))
	}
	if c.Listen == nil {
		str = append(str, "Listen: nil")
	} else {
		str = append(str, fmt.Sprintf("Listen: %d", *c.Listen))
	}
	return strings.Join(str, "\n")
}
