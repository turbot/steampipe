package control

import (
	"sync"

	"github.com/turbot/steampipe/pkg/control/controldisplay"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

type ExportData struct {
	ExecutionTree *controlexecute.ExecutionTree
	Targets       []controldisplay.CheckExportTarget
	ErrorsLock    *sync.Mutex
	Errors        []error
	WaitGroup     *sync.WaitGroup
}

func (e *ExportData) AddErrors(err []error) {
	e.ErrorsLock.Lock()
	e.Errors = append(e.Errors, err...)
	e.ErrorsLock.Unlock()
}
