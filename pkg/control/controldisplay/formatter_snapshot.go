package controldisplay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

type SnapshotFormatter struct {
	FormatterBase
}

func (f *SnapshotFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	snapshot, err := executionTreeToSnapshot(tree)
	if err != nil {
		return nil, err
	}

	// determine whether to indent the snapshot
	// TACTICAL: check in the context for contextKeyFormatterUse - if this is "export" then DO NOT indent
	var indent = true
	if formatterPurpose, ok := ctx.Value(contextKeyFormatterPurpose).(string); ok && formatterPurpose == formatterPurposeExport {
		indent = false
	}
	// strip unwanted fields from the snapshot
	snapshotStr, err := snapshot.AsStrippedJson(indent)
	if err != nil {
		return nil, err
	}

	res := strings.NewReader(fmt.Sprintf("%s\n", string(snapshotStr)))

	return res, nil
}

func (f *SnapshotFormatter) FileExtension() string {
	return constants.SnapshotExtension
}

func (f SnapshotFormatter) Name() string {
	return constants.OutputFormatSnapshot
}

func (f *SnapshotFormatter) Alias() string {
	return "sps"
}
