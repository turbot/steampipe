package controldisplay

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type CounterRendererOptions struct {
	AddLeadingSpace bool
}

type CounterRenderer struct {
	failedControls int
	totalControls  int

	maxFailedControls int
	maxTotalControls  int

	addLeadingSpace bool
}

func CounterRendererMinWidth() int {
	return 8
}

func NewCounterRenderer(failedControls, totalControls, maxFailedControls, maxTotalControls int, options CounterRendererOptions) *CounterRenderer {
	return &CounterRenderer{
		failedControls:    failedControls,
		totalControls:     totalControls,
		maxFailedControls: maxFailedControls,
		maxTotalControls:  maxTotalControls,
		addLeadingSpace:   options.AddLeadingSpace,
	}
}

/* Render returns the counter string in format "<failed> / <total>.
The alignment depends on the maximum failed and maximum total parameters, as the counters are aligned as follows:
"  3 /   123"
" 13 /    23"
"111 /   123"
"  1 /     4"
"  1 / 1,020"

minimum counter string is " n / m "

// NOTE: adds a trailing space
*/

const minimumCounterWidth = 7

func (r CounterRenderer) Render() string {
	p := message.NewPrinter(language.English)
	// get strings for fails and total - format with commas for thousands
	failedString := p.Sprintf("%d", r.failedControls)
	totalString := p.Sprintf("%d", r.totalControls)
	// get max strings - format with commas for thousands
	maxFailedString := p.Sprintf("%d", r.maxFailedControls)
	maxTotalString := p.Sprintf("%d", r.maxTotalControls)

	// calculate the width of the fails and total columns
	failedWidth := len(maxFailedString)
	if !r.addLeadingSpace {
		failedWidth = len(failedString)
	}
	totalWidth := len(maxTotalString)

	// build format string, specifying widths of failedString and totalString
	// this will generate a format string like: "%3s / %4s "
	// (adds a trailing space)
	formatString := fmt.Sprintf("%%%ds %%s %%%ds ", failedWidth, totalWidth)

	if r.failedControls == 0 {
		return fmt.Sprintf(formatString,
			ControlColors.CountZeroFail(failedString),
			ControlColors.CountZeroFailDivider("/"),
			ControlColors.CountTotalAllPassed(totalString))
	}

	str := fmt.Sprintf(formatString,
		ControlColors.CountFail(failedString),
		ControlColors.CountDivider("/"),
		ControlColors.CountTotal(totalString))
	return str
}
