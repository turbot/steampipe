package controldisplay

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
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
	builder := strings.Builder{}

	builder.WriteString(r.renderResult())
	builder.WriteString("\n")
	builder.WriteString(r.renderSummary())

	return builder.String()
}

func (r TableRenderer) renderSummary() string {
	// no need to render the summary when the dry-run flag is set
	if viper.GetBool(constants.ArgDryRun) {
		return ""
	}
	return NewSummaryRenderer(r.resultTree, r.width).Render()
}

func (r TableRenderer) renderResult() string {
	return NewGroupRenderer(r.resultTree.Root, nil, r.maxFailedControls, r.maxTotalControls, r.resultTree, r.width).Render()
}
