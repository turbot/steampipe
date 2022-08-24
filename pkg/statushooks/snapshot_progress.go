package statushooks

import "context"

type SnapshotProgress interface {
	UpdateRowCount(context.Context, int)
	UpdateErrorCount(context.Context, int)
}

func SnapshotError(ctx context.Context) {
	SnapshotProgressFromContext(ctx).UpdateErrorCount(ctx, 1)
}

func UpdateSnapshotProgress(ctx context.Context, completedRows int) {
	SnapshotProgressFromContext(ctx).UpdateRowCount(ctx, completedRows)
}
