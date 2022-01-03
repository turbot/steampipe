package cmd

import (
	"context"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/workspace"
)

type checkInitData struct {
	ctx       context.Context
	workspace *workspace.Workspace
	client    db_common.Client
	result    *db_common.InitResult
}
