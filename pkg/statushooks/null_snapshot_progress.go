package statushooks

import "context"

var NullProgress = &NullSnapshotProgress{}

type NullSnapshotProgress struct{}

func (*NullSnapshotProgress) UpdateRowCount(context.Context, int)   {}
func (*NullSnapshotProgress) UpdateErrorCount(context.Context, int) {}
