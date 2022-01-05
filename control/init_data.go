package control

import (
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/workspace"
)

type InitData struct {
	Workspace *workspace.Workspace
	Client    db_common.Client
	Result    *db_common.InitResult
}
