package error_helpers

import (
	"errors"
	"fmt"

	"github.com/turbot/pipe-fittings/v2/constants"
)

var MissingCloudTokenError = fmt.Errorf("Not authenticated for Turbot Pipes.\nPlease run %s or setup a token.", constants.Bold("steampipe login"))
var InvalidCloudTokenError = fmt.Errorf("Invalid token.\nPlease run %s or setup a token.", constants.Bold("steampipe login"))
var InvalidStateError = errors.New("invalid state")

// PluginSdkCompatibilityError is raised when aplugin is built using na incompatible sdk version
var PluginSdkCompatibilityError = fmt.Sprintf("plugins using SDK version < v4 are no longer supported. Upgrade by running %s", constants.Bold("steampipe plugin update --all"))
