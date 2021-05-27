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

func (j *CSVFormatter) Format(_ context.Context, tree *execute.ExecutionTree) (io.Reader, error) {
	renderer := newGroupCsvRenderer(tree.GetResultColumns())
	outBuffer := bytes.NewBufferString("")
	j.csvWriter = csv.NewWriter(outBuffer)
	data := renderer.Render(tree)
	j.csvWriter.Write(tree.GetResultColumns().AllColumns)
	j.csvWriter.WriteAll(data)
	return strings.NewReader(outBuffer.String()), nil
}
