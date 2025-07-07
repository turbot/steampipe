package runtime

import (
	"fmt"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

var (
	ExecutionID = helpers.GetMD5Hash(fmt.Sprintf("%d", time.Now().Nanosecond()))[:4]
)

var (
	// App name used by connections which issue user-initiated queries
	ClientConnectionAppName = fmt.Sprintf("%s_%s", constants.ClientConnectionAppNamePrefix, ExecutionID)

	// App name used for queries which support user-initiated queries (load schema, load connection state etc.)
	ClientSystemConnectionAppName = fmt.Sprintf("%s_%s", constants.ClientSystemConnectionAppNamePrefix, ExecutionID)

	// App name used for service related queries (plugin manager, refresh connection)
	ServiceConnectionAppName = fmt.Sprintf("%s_%s", constants.ServiceConnectionAppNamePrefix, ExecutionID)
)
