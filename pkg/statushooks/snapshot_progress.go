package statushooks

import "context"

type SnapshotProgress interface {
	UpdateRowCount(context.Context, int)
	UpdateErrorCount(context.Context, int)
	UploadComplete(ctx context.Context, snapshotUrl string)
	UploadError(ctx context.Context, err error)
}

func SnapshotError(ctx context.Context) {
	SnapshotProgressFromContext(ctx).UpdateErrorCount(ctx, 1)
}

func UpdateSnapshotProgress(ctx context.Context, completedRows int) {
	SnapshotProgressFromContext(ctx).UpdateRowCount(ctx, completedRows)
}

func UploadComplete(ctx context.Context, snapshotUrl string) {
	SnapshotProgressFromContext(ctx).UploadComplete(ctx, snapshotUrl)
}
func UploadError(ctx context.Context, err error) {
	SnapshotProgressFromContext(ctx).UploadError(ctx, err)
}
