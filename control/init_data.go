package control

import (
	"context"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/workspace"
)

type InitData struct {
	Ctx       context.Context
	Workspace *workspace.Workspace
	Client    db_common.Client
	Result    *db_common.InitResult
}
