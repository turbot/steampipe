package controldisplay

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

type SnapshotFormatter struct{}

func (f *SnapshotFormatter) Format(_ context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	snapshot, err := ExecutionTreeToSnapshot(tree)
	if err != nil {
		return nil, err
	}

	snapshotStr, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return nil, err
	}

	res := strings.NewReader(string(snapshotStr))

	return res, nil
}

func (f *SnapshotFormatter) FileExtension() string {
	return ".snapshot.json"
}

func (f SnapshotFormatter) Name() string {
	return constants.OutputFormatSnapshot
}
