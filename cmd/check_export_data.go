package cmd

import (
	"sync"

	"github.com/turbot/steampipe/control/controldisplay"
	"github.com/turbot/steampipe/control/controlexecute"
)

type checkExportData struct {
	executionTree *controlexecute.ExecutionTree
	exportFormats []controldisplay.CheckExportTarget
	errorsLock    *sync.Mutex
	errors        []error
	waitGroup     *sync.WaitGroup
}

func (e *checkExportData) addErrors(err []error) {
	e.errorsLock.Lock()
	e.errors = append(e.errors, err...)
	e.errorsLock.Unlock()
}
