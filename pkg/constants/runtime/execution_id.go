package runtime

import (
	"fmt"
	"time"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
)

var (
	ExecutionID     = helpers.GetMD5Hash(fmt.Sprintf("%d", time.Now().Nanosecond()))[:4]
	PgClientAppName = fmt.Sprintf("%s_%s", constants.AppName, ExecutionID)
)
