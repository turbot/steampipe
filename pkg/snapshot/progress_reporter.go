package snapshot

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
	"sync"
)

type SnapshotProgressReporter struct {
	rows            int
	errors          int
	nodeType        string
	name            string
	snapshotAddress string

	mut sync.Mutex
}

func NewSnapshotProgressReporter(target string, snapshotAddress string) *SnapshotProgressReporter {
	res := &SnapshotProgressReporter{
		name:            target,
		snapshotAddress: snapshotAddress,
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
	var rowString, errorString string
	if r.rows > 0 {
		rowString = fmt.Sprintf("%d %s returned, ", r.rows, utils.Pluralize("row", r.rows))
	}
	if r.errors > 0 {
		errorString = fmt.Sprintf("%d %s, ", r.errors, utils.Pluralize("error", r.errors))
	}

	message := fmt.Sprintf("Running %s, %s%spublishing snapshot to %s",
		r.name,
		rowString,
		errorString,
		r.snapshotAddress,
	)

	statushooks.SetStatus(ctx, message)
}
