package controldisplay

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
)

type SnapshotFormatter struct{}

func (j *SnapshotFormatter) Format(ctx context.Context, tree *controlexecute.ExecutionTree) (io.Reader, error) {
	var outputString = ""
	res := strings.NewReader(fmt.Sprintf("\n%s\n", outputString))
	return res, nil
}

func (j *SnapshotFormatter) FileExtension() string {
	return constants.JsonExtension
}

func (tf SnapshotFormatter) GetFormatName() string {
	return constants.OutputFormatSnapshot
}
