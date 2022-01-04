package runtime

import (
	"fmt"
	"time"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

var (
	ExecutionID     = utils.GetMD5Hash(fmt.Sprintf("%d", time.Now().Nanosecond()))
	PgClientAppName = fmt.Sprintf("%s_%s", constants.AppName, ExecutionID)
)
