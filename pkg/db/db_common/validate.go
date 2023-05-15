package db_common

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
)

func CheckReservedConnectionName(connectionName string) error {
	if strings.EqualFold(connectionName, "public") {
		return fmt.Errorf("'%s' is a reserved connection name", constants.Bold("public"))
	}
	if strings.HasPrefix(connectionName, "sp_") {
		return fmt.Errorf("connection name '%s' cannot start with '%s'", constants.Bold(connectionName), constants.Bold("sp_"))
	}
	return nil
}
