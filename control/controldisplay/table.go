package controldisplay

import (
	"bytes"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/turbot/steampipe/control/controlexecute"
)

type TableRenderer struct {
	resultTree *controlexecute.ExecutionTree

	// screen width
	width             int
	maxFailedControls int
	maxTotalControls  int
}

func NewTableRenderer(resultTree *controlexecute.ExecutionTree, width int) *TableRenderer {
	return &TableRenderer{
		resultTree:        resultTree,
		width:             width,
		maxFailedControls: resultTree.Root.Summary.Status.FailedCount(),
		maxTotalControls:  resultTree.Root.Summary.Status.TotalCount(),
	}
}

func (r TableRenderer) Render() string {
	// the buffer to put the output data in
	outbuf := bytes.NewBufferString("")

	outbuf.WriteString(r.renderSummary())
	outbuf.WriteString("\n")
	outbuf.WriteString(r.renderResult())

	return outbuf.String()
}

func (r TableRenderer) renderSummary() string {
	// the table
	t := table.NewWriter()
	t.SetStyle(table.StyleDefault)

	colConfigs := []table.ColumnConfig{}
	headers := make(table.Row, 5)

	for idx, column := range []string{"Alarm", "Ok", "Info", "Skip", "Error"} {
		headers[idx] = column
		colConfigs = append(colConfigs, table.ColumnConfig{
			Name:     column,
			Number:   idx + 1,
			WidthMin: (r.width / 5),
			WidthMax: (r.width / 5),
		})
	}
	t.SetColumnConfigs(colConfigs)
	t.AppendHeader(headers)

	t.AppendRow(table.Row{
		r.resultTree.Root.Summary.Status.Alarm,
		r.resultTree.Root.Summary.Status.Ok,
		r.resultTree.Root.Summary.Status.Info,
		r.resultTree.Root.Summary.Status.Skip,
		r.resultTree.Root.Summary.Status.Error,
	})
	return t.Render()
}

func (r TableRenderer) renderResult() string {
	return NewGroupRenderer(r.resultTree.Root, nil, r.maxFailedControls, r.maxTotalControls, r.resultTree, r.width).Render()
}
