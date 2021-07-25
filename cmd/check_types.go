package cmd

import (
	"context"
	"sync"

	"github.com/turbot/steampipe/db/db_common"

	"github.com/turbot/steampipe/control/controldisplay"
	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/workspace"
)

type checkInitData struct {
	ctx           context.Context
	workspace     *workspace.Workspace
	client        db_common.Client
	dbInitialised bool
	error         error
	warning       string
}

func (c *checkInitData) success() bool {
	return c.error == nil && c.warning == "" && c.ctx.Err() == nil
}

type exportData struct {
	executionTree *controlexecute.ExecutionTree
	exportFormats []controldisplay.CheckExportTarget
	errorsLock    *sync.Mutex
	errors        []error
	waitGroup     *sync.WaitGroup
}
