package cmd

import (
	"context"
	"sync"

	"github.com/turbot/steampipe/db/local_db"

	"github.com/turbot/steampipe/control/controldisplay"
	"github.com/turbot/steampipe/control/controlexecute"
	"github.com/turbot/steampipe/workspace"
)

type checkInitData struct {
	ctx           context.Context
	workspace     *workspace.Workspace
	client        *local_db.LocalClient
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
