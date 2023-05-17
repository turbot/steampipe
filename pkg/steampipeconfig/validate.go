package steampipeconfig

import (
	"fmt"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"strings"
)

func ValidateConnectionName(connectionName string) error {
	if helpers.StringSliceContains(constants.ReservedConnectionNames, connectionName) {
		return fmt.Errorf("'%s' is a reserved connection name", constants.Bold(connectionName))
	}
	if strings.HasPrefix(connectionName, constants.ReservedConnectionNamePrefix) {
		return fmt.Errorf("invalid connection name '%s' - connection names cannot start with '%s'", constants.Bold(connectionName), constants.Bold(constants.ReservedConnectionNamePrefix))
	}
	return nil
}
