package statushooks

import "context"

// NullProgress is an empty implementation of SnapshotProgress
var NullProgress = &NullSnapshotProgress{}

type NullSnapshotProgress struct{}

func (*NullSnapshotProgress) UpdateRowCount(context.Context, int)   {}
func (*NullSnapshotProgress) UpdateErrorCount(context.Context, int) {}
