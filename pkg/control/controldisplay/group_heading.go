package controldisplay

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
)

type GroupHeadingRenderer struct {
	title             string
	severity          string
	failedControls    int
	totalControls     int
	maxFailedControls int
	maxTotalControls  int
	// screen width
	width  int
	indent string
}

func NewGroupHeadingRenderer(title string, failed, total, maxFailed, maxTotal, width int, indent string) *GroupHeadingRenderer {
	return &GroupHeadingRenderer{
		title:             title,
		failedControls:    failed,
		totalControls:     total,
		maxFailedControls: maxFailed,
		maxTotalControls:  maxTotal,
		width:             width,
		indent:            indent,
	}
}

func (r GroupHeadingRenderer) Render() string {
	isDryRun := viper.GetBool(constants.ArgDryRun)

	if r.width <= 0 {
		// this should never happen, since the minimum width is set by the formatter
		log.Printf("[WARN] group heading renderer has width of %d\n", r.width)
		return ""
	}

	formattedIndent := fmt.Sprintf("%s", ControlColors.Indent(r.indent))
	indentWidth := helpers.PrintableLength(formattedIndent)

	// for a dry run we do not display the counters or graph
	var severityString, counterString, graphString string
	if !isDryRun {
		severityString = NewSeverityRenderer(r.severity).Render()
		counterString = NewCounterRenderer(
			r.failedControls,
			r.totalControls,
			r.maxFailedControls,
			r.maxTotalControls,
			CounterRendererOptions{
				AddLeadingSpace: true,
			},
		).Render()

		graphString = NewCounterGraphRenderer(
			r.failedControls,
			r.totalControls,
			r.maxTotalControls,
			CounterGraphRendererOptions{
				FailedColorFunc: ControlColors.CountGraphFail,
			},
		).Render()
	}
	severityWidth := helpers.PrintableLength(severityString)
	counterWidth := helpers.PrintableLength(counterString)
	graphWidth := helpers.PrintableLength(graphString)

	// figure out how much width we have available for the title
	availableWidth := r.width - counterWidth - graphWidth - severityWidth - indentWidth

	// now availableWidth is all we have - if it is not enough we need to truncate the title
	titleString := NewGroupTitleRenderer(r.title, availableWidth).Render()
	titleWidth := helpers.PrintableLength(titleString)

	// is there any room for a spacer
	spacerWidth := availableWidth - titleWidth
	var spacerString string
	if spacerWidth > 0 && !isDryRun {
		spacerString = NewSpacerRenderer(spacerWidth).Render()
	}

	// now put these all together
	str := fmt.Sprintf("%s%s%s%s%s%s", formattedIndent, titleString, spacerString, severityString, counterString, graphString)
	return str
}
