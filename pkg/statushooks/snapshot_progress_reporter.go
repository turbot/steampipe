package statushooks

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/utils"
	"strings"
	"sync"
)

// SnapshotProgressReporter is an implementation of SnapshotProgress
type SnapshotProgressReporter struct {
	rows     int
	errors   int
	nodeType string
	name     string
	mut      sync.Mutex
}

func NewSnapshotProgressReporter(target string) *SnapshotProgressReporter {
	res := &SnapshotProgressReporter{
		name: target,
	}
	return res
}

func (r *SnapshotProgressReporter) UpdateRowCount(ctx context.Context, rows int) {
	r.mut.Lock()
	defer r.mut.Unlock()

	r.rows += rows
	r.showProgress(ctx)
}
func (r *SnapshotProgressReporter) UpdateErrorCount(ctx context.Context, errors int) {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.errors += errors
	r.showProgress(ctx)
}

func (r *SnapshotProgressReporter) showProgress(ctx context.Context) {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("Running %s", r.name))
	if r.rows > 0 {
		msg.WriteString(fmt.Sprintf(", %d %s returned", r.rows, utils.Pluralize("row", r.rows)))
	}
	if r.errors > 0 {
		msg.WriteString(fmt.Sprintf(", %d %s, ", r.errors, utils.Pluralize("error", r.errors)))
	}

	SetStatus(ctx, msg.String())
}
