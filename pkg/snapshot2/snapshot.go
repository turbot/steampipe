package snapshot2

import (
	"context"

	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/turbot/steampipe/pkg/initialisation"
)

func GenerateSnapshot(ctx context.Context, target string, initData *initialisation.InitData, inputs map[string]any) (snapshot steampipeconfig.SteampipeSnapshot, err error) {
	snapshot = NewEmptySnapshot()
	return snapshot, nil
}

func NewEmptySnapshot() steampipeconfig.SteampipeSnapshot {
	return steampipeconfig.SteampipeSnapshot{
		SchemaVersion: "20221222",
		Inputs:        make(map[string]interface{}),
		Panels:        make(map[string]steampipeconfig.SnapshotPanel),
		Variables:     make(map[string]string),
	}
}
