package statushooks

import "context"

var NullProgress = &NullSnapshotProgress{}

type NullSnapshotProgress struct{}

func (*NullSnapshotProgress) UpdateRowCount(context.Context, int)    {}
func (*NullSnapshotProgress) UpdateErrorCount(context.Context, int)  {}
func (*NullSnapshotProgress) UploadComplete(context.Context, string) {}
func (*NullSnapshotProgress) UploadError(context.Context, error)     {}
