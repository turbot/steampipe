package controldisplay

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"strings"

	"github.com/turbot/steampipe/control/execute"
)

type CSVFormatter struct {
	csvWriter *csv.Writer
}

func (j *CSVFormatter) Format(ctx context.Context, tree *execute.ExecutionTree) (io.Reader, error) {
	outbuf := bytes.NewBufferString("")
	j.csvWriter = csv.NewWriter(outbuf)
	j.renderGroupRun(ctx, tree.Root, nil)
	j.csvWriter.Flush()
	if j.csvWriter.Error() != nil {
		return nil, j.csvWriter.Error()
	}
	return strings.NewReader(outbuf.String()), nil
}

func (j *CSVFormatter) renderGroupRun(ctx context.Context, groupRun *execute.ResultGroup, prepend []string) {
	record := append(prepend, groupRun.Title)

	j.csvWriter.Write(record)

	for _, group := range groupRun.Groups {
		j.renderGroupRun(ctx, group, record)
	}
	for _, ctrl := range groupRun.ControlRuns {
		j.renderControlRun(ctx, ctrl, record)
	}
}
func (j *CSVFormatter) renderControlRun(ctx context.Context, controlRun *execute.ControlRun, prepend []string) {
	thisPrepend := append(prepend, controlRun.Title)
	j.csvWriter.Write(thisPrepend)

	for _, row := range controlRun.Rows {
		record := append(thisPrepend, row.Status, row.Reason)

		for _, dimension := range row.Dimensions {
			record = append(record, dimension.Value)
		}

		j.csvWriter.Write(record)
	}

}
