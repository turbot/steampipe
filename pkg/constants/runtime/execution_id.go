package runtime

import (
	"fmt"
	"time"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

var (
	ExecutionID     = utils.GetMD5Hash(fmt.Sprintf("%d", time.Now().Nanosecond()))[:4]
	// PgClientAppName is unique identifier for this execution of Steampipe
	PgClientAppName = fmt.Sprintf("%s_%s", constants.AppName, ExecutionID)
)
