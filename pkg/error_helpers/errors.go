package error_helpers

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/constants"
)

var MissingCloudTokenError = fmt.Errorf("Not authenticated for Steampipe Cloud.\nPlease run %s or setup a token.", constants.Bold("steampipe login"))
