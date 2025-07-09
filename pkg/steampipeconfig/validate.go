package steampipeconfig

import (
	"fmt"
	"slices"
	"strings"

	"github.com/turbot/steampipe/v2/pkg/constants"
)

func ValidateConnectionName(connectionName string) error {
	if slices.Contains(constants.ReservedConnectionNames, connectionName) {
		return fmt.Errorf("'%s' is a reserved connection name", connectionName)
	}
	if strings.HasPrefix(connectionName, constants.ReservedConnectionNamePrefix) {
		return fmt.Errorf("invalid connection name '%s' - connection names cannot start with '%s'", connectionName, constants.ReservedConnectionNamePrefix)
	}
	return nil
}
